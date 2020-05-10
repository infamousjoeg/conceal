package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/infamousjoeg/conceal/pkg/conceal"
	"github.com/infamousjoeg/conceal/pkg/conceal/keychain"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	// Set command-line options
	account := flag.String("a", "", "Account name for secret to add")
	list := flag.Bool("l", false, "List of Summon accounts in keychain")
	remove := flag.Bool("r", false, "Remove account from keychain")
	version := flag.Bool("v", false, "Display current version")

	flag.Parse()

	// If an account is given but not set using the -a flag, set account pointer to it
	if *account == "" {
		*account = flag.Arg(0)
	}

	// If an account is not given and version or list is not requested, show help
	if *account == "" && !*version && !*list && !*remove {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// If remove is requested, remove secret from keychain
	if *account != "" && *remove {
		keychain.DeleteSecret(*account)
		os.Exit(1)
	}
	if flag.Arg(0) != "" && flag.Arg(1) == "-r" {
		keychain.DeleteSecret(flag.Arg(0))
		os.Exit(1)
	}

	// If list is requested, return list of accounts with `Summon` label
	if *list {
		accounts := keychain.ListSecrets()
		fmt.Println("The following Summon accounts are in keychain:")
		for account := range accounts {
			fmt.Println(accounts[account])
		}
		os.Exit(1)
	}

	// If version is requested, give it
	if *version {
		fmt.Printf("conceal v%s\n", conceal.FullVersionName)
		os.Exit(1)
	}

	// Get secret value from STDIN
	fmt.Println("Please enter the secret value: ")
	byteSecretVal, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalln("An error occurred trying to read password from " +
			"Stdin. Exiting...")
	}

	// Add secret and secret value to keychain
	keychain.AddSecret(*account, byteSecretVal)
}
