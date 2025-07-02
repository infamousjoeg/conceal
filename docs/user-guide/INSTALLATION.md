# Installation

The easiest way to try Conceal is to install from a package manager. Prebuilt
binaries are also available on the [releases page](https://github.com/infamousjoeg/conceal/releases).

## macOS with Homebrew

```bash
brew tap infamousjoeg/tap
brew install conceal
```

## Windows with Scoop

```powershell
scoop bucket add conceal https://github.com/infamousjoeg/scoop
scoop install conceal
```

## Linux or manual installation

Download the archive for your platform and place the `conceal` binary somewhere
in your `PATH`. On Linux install `libsecret-1` and run a compatible secret
service such as `gnome-keyring`.

## From source

```bash
go install github.com/infamousjoeg/conceal@latest
```
