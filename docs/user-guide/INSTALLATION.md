# Installation

## Homebrew (macOS)
```bash
brew tap infamousjoeg/tap
brew install conceal
```

## Windows Scoop
```powershell
scoop bucket add conceal https://github.com/infamousjoeg/scoop
scoop install conceal
```

## Binary releases
Download the archive for your platform from the [releases page](https://github.com/infamousjoeg/conceal/releases) and place the `conceal` executable somewhere in your `PATH`.

## From source
```bash
go install github.com/infamousjoeg/conceal@latest
```

On Linux ensure `libsecret-1` and a running secret service like `gnome-keyring` are available.
