# Architecture Overview

Conceal is a thin CLI wrapper around each operating system's credential store. A
single `keychain` package provides the same functions on macOS, Windows and
Linux.

```
+---------+      +------------+
|  CLI    +----->+  keychain  |
+---------+      +------------+
                      | platform implementations
```

Backends expose `AddSecret`, `GetSecret`, `UpdateSecret`, `DeleteSecret`,
`ListSecrets` and `SecretExists`. If a platform is unsupported the functions
return an informative error.

Secrets are encrypted at rest by the OS. On Linux we use `libsecret` over D-Bus
which requires a running keyring service like `gnome-keyring`.
