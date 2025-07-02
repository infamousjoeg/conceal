# Public API

The `keychain` package abstracts secret operations. Import it to embed Conceal
functionality in other Go programs.

```go
import "github.com/infamousjoeg/conceal/pkg/conceal/keychain"

err := keychain.AddSecret("db/pass", []byte("s3cr3t"))
```

Each backend implements the same functions. An error is returned if the platform
is unsupported or the secret service is unavailable.
