# Security Best Practices

* Limit local user access to secret stores.
* Rotate secrets regularly using `conceal update`.
* Prefer piping secrets via `conceal get --stdout` into commands rather than
storing them in scripts.
