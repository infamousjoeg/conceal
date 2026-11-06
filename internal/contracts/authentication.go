package contracts

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"slices"
	"strings"

	awscredentials "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/infamousjoeg/conceal/pkg/conceal/adapter"
)

// authenticate is deliberately host-owned. Adapters receive neither a secret
// store nor values, and signing cannot trigger credential discovery or API calls.
func (m Manager) authenticate(ctx context.Context, request *http.Request, plan adapter.Plan, references map[string]string) ([][]byte, error) {
	var buffers [][]byte
	if m.ReadSecret == nil {
		return buffers, fmt.Errorf("contract credential store unavailable")
	}
	roles := make([]string, 0, len(references))
	for role := range references {
		roles = append(roles, role)
	}
	slices.Sort(roles)
	values := make(map[string][]byte)
	for _, role := range roles {
		value, err := m.ReadSecret(references[role])
		buffers = append(buffers, value)
		if err != nil {
			return buffers, fmt.Errorf("credential role %q is unavailable", role)
		}
		if len(value) == 0 || strings.ContainsAny(string(value), "\x00\r\n") {
			return buffers, fmt.Errorf("credential role %q contains an invalid value", role)
		}
		values[role] = value
	}
	switch plan.Authentication.Scheme {
	case "bearer":
		if len(values["token"]) == 0 {
			return buffers, fmt.Errorf("token credential is required")
		}
		request.Header.Set("Authorization", "Bearer "+string(values["token"]))
	case "aws-sigv4":
		if len(values["access_key_id"]) == 0 || len(values["secret_access_key"]) == 0 {
			return buffers, fmt.Errorf("AWS access key ID and secret access key are required")
		}
		payloadHash := sha256.Sum256(plan.HTTP.Body)
		err := v4.NewSigner().SignHTTP(ctx, awscredentials.Credentials{
			AccessKeyID: string(values["access_key_id"]), SecretAccessKey: string(values["secret_access_key"]),
			SessionToken: string(values["session_token"]),
		}, request, hex.EncodeToString(payloadHash[:]), plan.Authentication.Service, plan.Authentication.Region, m.now())
		if err != nil {
			return buffers, fmt.Errorf("AWS request signing failed")
		}
	default:
		return buffers, fmt.Errorf("unsupported authentication scheme")
	}
	return buffers, nil
}
