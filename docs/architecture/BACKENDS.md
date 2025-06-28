# Secret Store Backends

## macOS Keychain
The macOS backend uses the `go-keychain` library. Items are stored as generic
passwords scoped to the user. Each secret's label is the secret ID used by
Conceal.

```go
item := keychain.NewItem()
item.SetSecClass(keychain.SecClassGenericPassword)
item.SetService("conceal")
item.SetAccount(secretID)
item.SetLabel(secretID)
item.SetData(secret)
err := keychain.AddItem(item)
```

## Windows Credential Manager
The Windows backend relies on the `wincred` library to interact with Credential
Manager. Secrets are stored as generic credentials.

```go
cred := wincred.NewGenericCredential(secretID)
cred.CredentialBlob = secret
cred.Comment = "summon"
cred.Write()
```

## Linux libsecret
Linux support uses `go-libsecret` which communicates with a D-Bus secret
service such as `gnome-keyring`.

```go
svc, _ := libsecret.NewService()
session, _ := svc.Open()
col := svc.Collections()[0]
sec := libsecret.NewSecret(session, nil, secret, "text/plain")
col.CreateItem(secretID, sec, false)
```

If no secret service is available, errors are wrapped with hints so the user can
start `gnome-keyring` or an alternative service.
