# Public API

The main package exposes a `keychain` package which abstracts secret store
operations.

```go
import "github.com/infamousjoeg/conceal/pkg/conceal/keychain"

err := keychain.AddSecret("db/pass", []byte("s3cr3t"))
```

Each backend implements the same functions. Errors are returned if the platform
is unsupported or the secret service is unavailable.
