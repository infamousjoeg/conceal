# Basic Usage

```bash
# set a secret
echo "demo" | conceal set example/secret

# update it
echo "new" | conceal update example/secret

# print the value
conceal get example/secret --stdout

# remove it
conceal unset example/secret
```
