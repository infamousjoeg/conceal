# CI/CD Patterns

Use conceal to inject secrets into build jobs:
```yaml
- name: Fetch secret
  run: echo "DB_PASS=$(conceal get db/password --stdout)" >> $GITHUB_ENV
```
