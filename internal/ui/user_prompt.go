package ui

import (
	"log"

	"github.com/AlecAivazis/survey/v2"
)

func user_prompt() {
	var action string
	prompt := &survey.Select{
		Message: "Choose an option:",
		Options: []string{"Login", "Sign Up"},
		Default: "Login",
	}

	err := survey.AskOne(prompt, &action)
	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}

	switch action {
	case "Login":
		login()
	case "Sign Up":
		SignUpCmd.Run(nil, nil)
	}
}
