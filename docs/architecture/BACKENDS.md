# Secret Store Backends

## macOS Keychain

Uses the `go-keychain` library. Items are stored as generic passwords and scoped
to the user.

```go
item := keychain.NewItem()
item.SetSecClass(keychain.SecClassGenericPassword)
item.SetService("conceal")
item.SetAccount(secretID)
item.SetLabel(secretID)
item.SetData(secret)
keychain.AddItem(item)
```

## Windows Credential Manager

Relies on the `wincred` library to store generic credentials.

```go
cred := wincred.NewGenericCredential(secretID)
cred.CredentialBlob = secret
cred.Write()
```

## Linux libsecret

Uses `go-libsecret` over D-Bus. A secret service like `gnome-keyring` must be
running.

```go
svc, _ := libsecret.NewService()
session, _ := svc.Open()
col := svc.Collections()[0]
sec := libsecret.NewSecret(session, nil, secret, "text/plain")
col.CreateItem(secretID, sec, false)
```
