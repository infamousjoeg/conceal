package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/infamousjoeg/conceal/pkg/conceal"
	"github.com/infamousjoeg/conceal/pkg/conceal/keychain"
)

func main() {
	// Set command-line options
	account := flag.String("a", "", "Account name for secret to add")
	list := flag.Bool("l", false, "List of Summon accounts in keychain")
	remove := flag.Bool("r", false, "Remove account from keychain")
	version := flag.Bool("v", false, "Display current version")

	flag.Parse()

	// If an account is not given and version is not requested, error
	if flag.Arg(0) == "" && *version == false && *list == false {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// If an account is given but not set using the -a flag, continue
	if flag.Arg(0) != "" && *account == "" {
		*account = flag.Arg(0)
	}

	// If list is requested, return list of accounts with `Summon` label
	if *list == true {
		accounts := keychain.ListSecrets()
		fmt.Println("The following Summon accounts are in keychain:")
		for account := range accounts {
			fmt.Println(accounts[account])
		}
		os.Exit(1)
	}

	// If version is requested, give it
	if *version == true {
		fmt.Printf("conceal v%s\n", conceal.FullVersionName)
		os.Exit(1)
	}

	// If remove is requested, remove secret from keychain
	if *account != "" && *remove == true {
		keychain.DeleteSecret(*account)
		os.Exit(1)
	}

	// Add secret and secret value to keychain
	keychain.AddSecret(*account)
}
