package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// askConfirmation prompts the user for yes/no confirmation
func askConfirmation(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", prompt)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
