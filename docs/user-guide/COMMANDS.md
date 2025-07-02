# Command Reference

Below are the most common commands. Run `conceal <command> --help` for full
usage.

## conceal set

Store a new secret. Pipe the value or type when prompted.

```bash
echo "mypwd" | conceal set db/password
```

## conceal get

Retrieve a secret. The value is copied to the clipboard for 15 seconds. Use
`--stdout` to print it instead.

```bash
conceal get db/password --stdout | some-command
```

## conceal update

Overwrite an existing secret.

```bash
echo "new" | conceal update db/password
```

## conceal unset

Remove a secret from the store.

## conceal list

List stored secret IDs.

## conceal summon install

Enable Conceal as a Summon provider so YAML entries like `!var db/password`
work out of the box.
