# Building Conceal

Conceal requires Go 1.21 or newer. After cloning the repository:

```bash
go mod tidy
go build ./...
```

To cross‑compile for another platform:

```bash
GOOS=windows GOARCH=amd64 go build -o conceal.exe
```
