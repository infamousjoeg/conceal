module github.com/infamousjoeg/conceal

go 1.23.0

toolchain go1.23.8

require (
	github.com/atotto/clipboard v0.1.4
	github.com/aws/aws-sdk-go-v2 v1.33.0
	github.com/aws/aws-sdk-go-v2/config v1.29.0
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.26.0
	github.com/danieljoos/wincred v1.2.2
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-plugin v1.6.3
	github.com/keybase/go-keychain v0.0.0-20231219164618-57a3676c3af6
	github.com/spf13/cobra v1.8.0
	golang.org/x/term v0.32.0
)

replace golang.org/x/net => github.com/golang/net v0.41.0

replace golang.org/x/sys => github.com/golang/sys v0.33.0

replace google.golang.org/grpc => github.com/grpc/grpc-go v1.73.0

replace google.golang.org/protobuf => github.com/protocolbuffers/protobuf-go v1.36.6

replace google.golang.org/genproto => github.com/googleapis/go-genproto v0.0.0-20250603155806-513f23925822

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.17.53 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.24 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.28 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.28 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.8 // indirect
	github.com/aws/smithy-go v1.22.1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/godbus/dbus v4.1.0+incompatible // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/grpc v1.58.3 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
