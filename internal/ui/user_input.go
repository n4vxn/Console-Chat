package ui

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/n4vxn/Console-Chat/internal/db"
	"github.com/n4vxn/Console-Chat/types"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

func login() {
	app := tview.NewApplication()
	form := tview.NewForm().
		AddInputField("Username", "", 20, nil, nil).
		AddPasswordField("Password", "", 10, '*', nil).
		AddButton("Login", nil).
		AddButton("Quit", func() {
			app.Stop()
		})
	form.SetBorder(true).SetTitle("Enter some data").SetTitleAlign(tview.AlignLeft)
	if err := app.SetRoot(form, true).EnableMouse(true).EnablePaste(true).Run(); err != nil {
		panic(err)
	}
}

var SignUpCmd = &cobra.Command{
	Use:   "signup",
	Short: "Sign up a new user.",
	Run: func(cmd *cobra.Command, args []string) {
		var username, password string

		survey.AskOne(&survey.Input{Message: "Enter Username:"}, &username)
		survey.AskOne(&survey.Input{Message: "Enter Password:"}, &password)

		user := types.Users{
			Username: username,
			Password: password,
		}

		db.SaveUsers(&user)
	},
}
