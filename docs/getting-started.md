# Get something working

[Back to Conceal](../README.md)

Choose a first task:

- [Run a Python app with a stored key](#run-your-first-app).
- [Review a GitHub action without a token](actions.md#try-it-without-a-token).
- [Try the desktop approval inbox](desktop.md) or [review an S3 backup listing](aws.md).
- [Use an existing Summon workflow](#use-your-existing-summon-workflow).

The examples use made-up values. You do not need a real API account to try them.

## Install the release candidate

**The `run` and `action` commands require 4.3.** Choose one of these candidate installation methods before following the examples. Homebrew and Chocolatey currently serve the older stable version.

### Download a candidate build

Open a successful run of the [release packaging workflow](https://github.com/infamousjoeg/conceal/actions/workflows/test-release.yml) for `release/v4.3.0-rc.1`, then download its `release-candidate-packages` artifact. GitHub requires you to sign in to download workflow artifacts.

Extract that artifact, then choose the archive for your system:

| System | Archive |
| --- | --- |
| macOS, Apple Silicon | `conceal_Darwin_arm64.tar.gz` |
| macOS, Intel | `conceal_Darwin_x86_64.tar.gz` |
| Windows, x64 | `conceal_Windows_x86_64.zip` |
| Windows, ARM64 | `conceal_Windows_arm64.zip` |
| Linux | `conceal_Linux_x86_64.tar.gz`, `conceal_Linux_arm64.tar.gz`, or `conceal_Linux_i386.tar.gz` |

Compare the archive's SHA-256 hash with its entry in `checksums.txt` before extracting it. Use `Get-FileHash <archive> -Algorithm SHA256` in PowerShell, `shasum -a 256 <archive>` on macOS, or `sha256sum <archive>` on Linux. Replace `<archive>` with the downloaded filename.

Extract the matching archive into a directory you control. It includes the executable, these guides, and the examples. Candidate macOS binaries require macOS 13+ and are not Developer ID signed or notarized. Windows requires Windows 10+ or Server 2016+; Linux requires a session that permits user keyrings.

For this terminal session, prepend the extracted directory to `PATH`. From inside that directory:

```powershell
# Windows PowerShell
$env:PATH = "$((Get-Location).Path);$env:PATH"
conceal version
```

```sh
# macOS / Linux
export PATH="$PWD:$PATH"
conceal version
```

The version should start with `conceal v4.3.0-rc.1`. If it does not, check which executable your terminal selected: `Get-Command conceal` in PowerShell or `command -v conceal` on macOS/Linux.

### Build the candidate yourself

Use Go 1.26 or newer; the release toolchain is Go 1.27.3. macOS additionally needs Xcode Command Line Tools and CGO enabled for the Security framework.

```sh
git clone --branch release/v4.3.0-rc.1 https://github.com/infamousjoeg/conceal.git
cd conceal
go build -trimpath -o bin/ .
```

From that checkout, select the new executable for this terminal:

```powershell
# Windows PowerShell
$env:PATH = "$(Join-Path (Get-Location).Path 'bin');$env:PATH"
conceal version
```

```sh
# macOS / Linux
export PATH="$PWD/bin:$PATH"
conceal version
```

Check for `conceal v4.3.0-rc.1` before continuing. These `PATH` changes last only for the current terminal.

### Looking for the stable release?

Use [GitHub Releases](https://github.com/infamousjoeg/conceal/releases), `brew tap infamousjoeg/tap` followed by `brew install conceal`, or `choco install conceal`. Use the documentation shipped with that version; the candidate examples below require 4.3.

## Run your first app

You need Conceal 4.3 and Python. Use a terminal where `conceal version` reports the candidate. In the commands below, use `python3` instead of `python` if that is how Python is installed on your system.

Create a new empty folder for this exercise and open it in that terminal. The example files belong there, separate from an existing project.

### 1. Store a made-up key

```sh
conceal set demo/python-key
```

At the hidden prompt, type `demo-value-not-a-real-key` and press Enter. Nothing appears while you type; that is expected. `demo/python-key` is a label you choose, not a file path.

If that name already exists, choose a new demo name and use it throughout this example. Do not overwrite an existing key.

### 2. Make a tiny app

Save this as `demo.py` in your working directory:

```python
import os

if not os.environ.get("API_TOKEN"):
    raise SystemExit("No key received. Launch this script with conceal run.")

print("Key received. Ready to call an API.")
```

### 3. Launch it with the key

```sh
conceal run --env API_TOKEN=demo/python-key -- python demo.py
```

Expected output:

```text
Key received. Ready to call an API.
```

`API_TOKEN` is the environment variable your app reads. `demo/python-key` is the stored label Conceal loads. The key itself is absent from the source file and command arguments.

**Your app receives the real value.** Only launch programs you trust this way. Conceal masks exact key values in terminal output by default, but a program can still write, transform, or send the key elsewhere.

### 4. Clean up the demo

```sh
conceal unset demo/python-key
```

Delete your `demo.py` when you are finished. Once this works, use your own app's environment variable name and a new label for a real key.

### If you already use `.env` files

Your code can keep reading `os.environ["API_TOKEN"]`. Store the value with `conceal set`, then launch the app with the matching `--env` mapping. Conceal does not import or edit `.env` files.

Python's [`load_dotenv()`](https://github.com/theskumar/python-dotenv#variable-expansion) preserves an existing environment value by default. `load_dotenv(override=True)` can replace the value Conceal supplies. Check your framework's precedence rules, verify the app works, then remove the old key from your own `.env` file if you no longer need it there.

For notebooks, launch the notebook server through Conceal so newly started kernels can inherit its environment:

```sh
conceal run --env API_TOKEN=my-app/api-key -- jupyter lab
```

This does not update a server or kernel that is already running.

### If the first run fails

| Message or symptom | Next step |
| --- | --- |
| `run` is not recognized as a command | Check `conceal version` and your `PATH`; an older stable executable may be selected. |
| Python cannot be found | Try `python3`, or install Python and reopen the terminal. |
| The stored name is missing | Run `conceal check --json demo/python-key` using the same account and session where you stored it. |
| The credential backend is unavailable | Check [platform requirements](reference.md#platforms). Linux containers may block `keyctl`. |
| The script prints “No key received” | Use the complete `conceal run` command, with `--env API_TOKEN=...` and `--` before Python. |

## Use your existing Summon workflow

Keep [Summon](https://github.com/cyberark/summon) and your `secrets.yml` files. Conceal supplies the local key values.

These steps assume Summon and Python are installed and the candidate is on `PATH`. Use `python3` where needed.

Start in a **new empty directory** for this exercise. Keep your existing project's `secrets.yml`; the demo below creates its own file.

1. Store a made-up value with `conceal set demo/summon-key`. Type `demo-value-not-a-real-key` at the hidden prompt. If the label is already in use, choose another and update the mapping below.
2. Install the native provider into a writable directory:

   ```sh
   conceal summon install --dir ./providers
   ```

3. Save this as `secrets.yml` in the current directory:

   ```yaml
   API_TOKEN: !var demo/summon-key
   ```

   `!var` tells Summon to ask its provider for the value stored under that label.

4. Run this nonsecret check:

   ```powershell
   # Windows PowerShell
   summon --provider .\providers\conceal_summon.exe python -c "import os; print('Key received' if os.environ.get('API_TOKEN') else 'No key received')"
   ```

   ```sh
   # macOS / Linux
   summon --provider ./providers/conceal_summon python3 -c "import os; print('Key received' if os.environ.get('API_TOKEN') else 'No key received')"
   ```

   Expected output: `Key received`.

5. Remove the demo key with `conceal unset demo/summon-key`. Remove the demo `secrets.yml` if you created it just for this exercise.

For your existing projects, point Summon at this provider and map `!var` entries to your Conceal labels. The launched program receives the real values, as it does with other Summon providers. [Provider locations, upgrades, and protocol details](reference.md#summon-provider).
