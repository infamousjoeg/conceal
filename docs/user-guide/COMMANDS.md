# Command Reference

## conceal set
Store a new secret. If the secret already exists, use `conceal update`.
```bash
echo "mypwd" | conceal set db/password
```

## conceal get
Retrieve a secret. By default the value is copied to the clipboard for 15 seconds.
Use `--stdout` to print it.
```bash
conceal get db/password --stdout | some-command
```

## conceal update
Overwrite an existing secret.
```bash
echo "new" | conceal update db/password
```

## conceal unset
Delete a secret from the store.

## conceal list
List stored secret IDs.

## conceal summon install
Install Conceal as a Summon provider so Summon can fetch secrets using
`!var`: `DB_PASSWORD: !var db/password`
