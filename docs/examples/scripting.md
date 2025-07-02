# Scripting

Conceal plays nicely with shell scripts. Use the `--stdout` flag to pass secrets
into other commands.

```bash
secret=$(conceal get api/key --stdout)
curl -H "Authorization: $secret" https://example.com
```
