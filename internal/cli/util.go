package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Colors
var noColor = lipgloss.NoColor{}
var errorRed = lipgloss.Color("124")
var okGreen = lipgloss.Color("40")
var warnYellow = lipgloss.Color("214")

var errorStyle = lipgloss.NewStyle().Bold(true).Foreground(errorRed).Background(noColor)
var successStyle = lipgloss.NewStyle().Foreground(okGreen).Background(noColor)
var warnStyle = lipgloss.NewStyle().Foreground(warnYellow)

func reportErrorAndDieS(msg string) {
	fmt.Fprintln(os.Stderr, errorStyle.Render(msg))
	os.Exit(1)
}

func reportErrorAndDie(err error) {
	fmt.Fprintln(os.Stderr, errorStyle.Render(err.Error()))
	os.Exit(1)
}

func reportErrorAndDieP(prefix string, err error) {
	fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("%s: %s", prefix, err.Error())))
	os.Exit(1)
}

func reportWarning(msg string) {
	fmt.Fprintln(os.Stderr, warnStyle.Render(msg))
}

func reportSuccess(msg string) {
	fmt.Fprintln(os.Stdout, successStyle.Render(msg))
}

func askForConfirmation(msg, confirmation string) bool {
	var response string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stdout, msg+" ")
		response, _ = r.ReadString('\n')
		response = strings.TrimSuffix(response, "\n")
		if response != "" {
			break
		}
	}
	return response == confirmation
}

func editFile(filePath string) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		for _, e := range editors {
			if editorPath, err := exec.LookPath(e); err == nil {
				editor = editorPath
				break
			}
		}
	}
	if editor == "" {
		reportErrorAndDieS("no suitable editor found")
	}
	// Split this in case the user has set $EDITOR to something like "emacsclient -n"
	chunks := strings.Split(editor, " ")
	chunks = append(chunks, filePath)
	toExec := exec.Command(chunks[0], chunks[1:]...)
	toExec.Stderr = os.Stderr
	toExec.Stdin = os.Stdin
	toExec.Stdout = os.Stdout
	toExec.Run()
}
