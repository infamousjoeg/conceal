package conceal

import (
	"fmt"
	"log"
)

// PrintSuccess is a function that prints a success message to the console. The message is in green and includes a checkmark emoji.
func PrintSuccess(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Printf("\033[32m✅ %s\033[0m\n", message)
}

// PrintFailure is a function that prints a failure message to the console. The message is in red and includes a crossmark emoji.
func PrintFailure(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Printf("\033[31m❌ %s\033[0m\n", message)
}

// PrintInfo is a function that prints an informational message to the console. The message is in cyan and includes an information emoji.
func PrintInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Printf("\033[36mℹ️ %s\033[0m\n", message)
}

// PrintError is a function that prints an error message to the console. The message is in red and includes an error emoji.
func PrintError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Fatalf("\033[31m🚨 %s\033[0m\n", message)
}

// GetSecretName is a function that prompts the user to input a secret name. If the user does not provide a secret name, the function will prompt the user to input a secret name until a valid secret name is provided.
func GetSecretName(args []string) string {
	var secretName string
	if len(args) == 0 {
		for secretName == "" {
			fmt.Println("Please enter the secret name: ")
			if _, err := fmt.Scanln(&secretName); err != nil {
				log.Println("input error:", err)
			}
		}
	} else {
		secretName = args[0]
	}
	return secretName
}
