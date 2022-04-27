package shell

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/acarl005/stripansi"
	"github.com/fatih/color"
	"log"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"
)

var cwd = "./"

func FormatInBox(message string) string {
	c := color.New(color.FgCyan)
	const (
		horizontal = "─"
		prefix     = "│  "
		suffix     = "  │"
	)
	parts := strings.Split(message, "\n")
	width := longestString(parts)
	horizontalLine := strings.Repeat(horizontal, width+stringWidth(prefix)+stringWidth(suffix)-2)
	out := c.Sprint("┌" + horizontalLine + "┐\n")
	for _, part := range parts {
		padding := strings.Repeat(" ", width-stringWidth(part))
		out += c.Sprint(prefix) + part + padding + c.Sprint(suffix) + "\n"
	}
	out += c.Sprint("└" + horizontalLine + "┘")
	return out
}

func RunCommand(name string, arg ...string) (string, error) {
	output, err := RunCommandQuiet(name, arg...)
	println(output)
	PrintCommandError(err)
	return output, err
}

func RunCommandQuiet(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	cmd.Dir = cwd
	println(cmd.String())
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

func PrintCommandError(err error) {
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			println(string(exitErr.Stderr))
		}
	}
}

func SetCwd(newCwd string) {
	cwd = newCwd
}

func stringWidth(s string) int {
	return utf8.RuneCountInString(stripansi.Strip(s))
}

func longestString(options []string) int {
	best := 0
	for _, option := range options {
		width := stringWidth(option)
		if width > best {
			best = width
		}
	}
	return best
}

func AskOne(p survey.Prompt, response interface{}, opts ...survey.AskOpt) {
	err := survey.AskOne(p, response, opts...)
	if err != nil {
		if err == terminal.InterruptErr {
			os.Exit(1)
		}
		log.Fatal(err)
	}
}
