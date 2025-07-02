# CI/CD Patterns

Use Conceal in your pipeline to retrieve secrets at runtime.

```bash
export SUMMON_PROVIDER=conceal
conceal get db/password --stdout | deploy-script
```

For Docker builds, copy the binary into your image and call it during setup.
