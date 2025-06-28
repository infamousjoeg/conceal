# Scripting Example

```bash
secret=$(conceal get service/token --stdout)
curl -H "Authorization: Bearer $secret" https://example.com
```
