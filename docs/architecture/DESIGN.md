# Architecture Overview

Conceal uses a thin abstraction layer over each operating system's credential
store. The CLI commands call into the `keychain` package which provides
implementations for macOS (Keychain), Windows (Credential Manager), and Linux
(libsecret).

```
+---------+      +------------+
|   CLI   +----->+  keychain  |
+---------+      +------------+
                     | platform implementations
```

Each backend exposes the same functions: `AddSecret`, `GetSecret`,
`UpdateSecret`, `DeleteSecret`, `ListSecrets`, and `SecretExists`. Unsupported
platforms are handled by `keychain_other.go` which returns informative errors.

Secrets are stored using the OS facilities which handle encryption at rest.
Errors from D-Bus or libsecret on Linux are wrapped to guide the user to install
a compatible secret service.
