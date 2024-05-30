package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/infamousjoeg/conceal/pkg/conceal"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Summon provider wrapper",
	Long: `This command creates a wrapper script for using Conceal as a Summon provider.
    
    Example Usage:
    $ conceal summon install`,
	Run: func(cmd *cobra.Command, args []string) {
		installWrapper()
	},
}

func installWrapper() {
	// Define the script content
	scriptContent := `#!/bin/bash

    # Check if the correct number of arguments are provided
    if [ "$#" -ne 1 ]; then
        echo "Usage: $0 <secret-id>"
        exit 1
    fi

    # Call the conceal binary with the get argument and the provided secret ID
    conceal summon show "$1"
    `

	// Find the full path of the summon executable
	summonPath, err := exec.LookPath("summon")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding Summon: %v\n", err)
		conceal.PrintInfo("Make sure Summon is installed and available in your PATH.")
		os.Exit(1)
	}

	// Get the directory where the summon executable is located
	summonDir := filepath.Dir(summonPath)
	providersPath := filepath.Join(summonDir, "Providers")
	scriptFilePath := filepath.Join(providersPath, "conceal_summon")

	// Create the Providers directory
	err = os.MkdirAll(providersPath, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Providers directory: %v\n", err)
		os.Exit(1)
	}

	// Write the script content to the file
	err = os.WriteFile(scriptFilePath, []byte(scriptContent), 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating wrapper script: %v\n", err)
		os.Exit(1)
	}

	conceal.PrintSuccess("Wrapper script 'conceal_summon' created successfully.")
	conceal.PrintInfo("To use: summon --provider conceal_summon ...")
}

func init() {
	summonCmd.AddCommand(installCmd)
}
