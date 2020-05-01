# Conceal <!-- omit in toc -->

Conceal is a command-line utility that eases the interaction between developer and OSX Keychain Access and Linux Keyring. It is the open-source companion to [Summon](https://cyberark.github.io/summon) as every secret added using this tool into Keychain/Keyring is added using Summon-compliant formatting.

## Table of Contents <!-- omit in toc -->
- [Requirements](#requirements)
- [Installation](#installation)
- [Usage](#usage)
  - [Add a secret](#add-a-secret)
  - [List Summon secrets](#list-summon-secrets)
  - [Remove a secret](#remove-a-secret)
  - [Display Help](#display-help)
  - [Display Version](#display-version)
- [Maintainer](#maintainer)
- [License](#license)

## Requirements

* MacOS or Linux

_Windows Credential Manager support coming soon._

## Installation

1. Download the latest release available at [GitHub Releases](https://github.com/infamousjoeg/go-conceal/releases).
2. Move the `conceal` executable file to a directory in your `PATH`. (I use `~/bin`.)
3. In Terminal, run the following command to make sure it's in your `PATH`: \
   `$ conceal -v`

## Usage

### Add a secret

`$ conceal -a dockerhub/token`

To add a secret to Keychain/Keyring, call `conceal` and use the `-a` argument to pass the account name to add. You will be immediately prompted to provide a secret value in a secure manner.

### List Summon secrets

`$ conceal -l`

To list all secrets associated with Summon in Keychain/Keyring, call `conceal` and use the `-l` argument to list all accounts present.

To filter the list further, pipe to `grep` like this `$ conceal -l | grep dockerhub/`.

### Remove a secret

`$ conceal -a dockerhub/token -r`

To remove a secret that was added for Summon, call `conceal` and use the `-a` argument to pass the account name to remove. The additional `-r` argument tells `conceal` to remove the secret instead of add.

### Display Help

`$ conceal`

To display the help message, just call `conceal` with no arguments.

### Display Version

`$ conceal -v`

To display the current version, call `conceal` with the `-v` argument.

## Maintainer

[@infamousjoeg](https://github.com/infamousjoeg)

[![Buy me a coffee][buymeacoffee-shield]][buymeacoffee]

[buymeacoffee]: https://www.buymeacoffee.com/infamousjoeg
[buymeacoffee-shield]: https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png

## Contributions

Pull Requests are currently being accepted.  Please read and follow the guidelines laid out in [](CONTRIBUTING.md).

## License

[Apache 2.0](LICENSE)
