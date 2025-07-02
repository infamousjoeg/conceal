# Summon Integration Example

```bash
conceal summon install
summon --yaml env.yml myapp
```

`env.yml`:
```yaml
DB_USER: !var db/user
DB_PASS: !var db/pass
```
