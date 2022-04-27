package commands

import "C"
import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/acarl005/stripansi"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"
)

var cwd = "./"

var PickCmd = cli.Command{
	Name:                   "pick",
	Usage:                  "pick <upstream branch>",
	Aliases:                []string{"p"},
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "skip-fetch"},
		&cli.BoolFlag{Name: "no-pr"},
		&cli.StringFlag{
			Name:    "branch",
			Aliases: []string{"b"},
		},
		&cli.StringSliceFlag{
			Name: "commits",
		},
		&cli.StringFlag{
			Name:  "pull-remote",
			Value: "upstream",
		},
		&cli.StringFlag{
			Name:  "push-remote",
			Value: "origin",
		},
	},
	BashComplete: func(context *cli.Context) {

	},
	Action: func(context *cli.Context) error {
		pullRemote := context.String("pull-remote")
		pushRemote := context.String("push-remote")

		upstreamBranch := context.Args().First()
		if len(upstreamBranch) < 1 {
			return cli.Exit("no upstream branch provided", 1)
		}
		upstreamRef := pullRemote + "/" + upstreamBranch

		var pickedCommits []string
		var pickedCommitShas []string
		if context.IsSet("commits") {
			pickedCommitShas = context.StringSlice("commits")
			pickedCommitShas, err := verifyCommits(pickedCommitShas)
			if err != nil {
				return err
			}
			pickedCommits, err = getCommits(strings.Join(pickedCommitShas, " "))
			if err != nil {
				return err
			}
		} else {
			recentCommits, err := getCommits(upstreamRef + "..")
			if err != nil {
				return err
			}

			askOne(&survey.MultiSelect{
				Message: "Which commits do you want to pick?",
				Options: recentCommits,
			}, &pickedCommits, survey.WithKeepFilter(true), survey.WithValidator(survey.MinItems(1)))

			for _, commit := range pickedCommits {
				commitSha := strings.SplitN(commit, " ", 2)[0]
				pickedCommitShas = append(pickedCommitShas, commitSha)
			}
		}

		if len(pickedCommitShas) < 1 {
			return cli.Exit("No commits chosen", 1)
		}

		var branchName string
		branchValidator := survey.MinLength(3)
		if context.IsSet("branch") {
			branchName = context.String("branch")
		} else {
			askOne(&survey.Input{
				Message: "Branch name:",
			}, &branchName, survey.WithValidator(branchValidator))
		}
		err := branchValidator(branchName)
		if err != nil {
			return err
		}

		infoSign := color.HiGreenString("ℹ ")
		var commitLines string
		for _, commit := range pickedCommits {
			commitLines += "\n" + "   └▷ " + color.CyanString(commit)
		}
		lines := []string{
			infoSign + "About to cherry pick commit(s):" + commitLines,
			infoSign + "Base branch: " + color.CyanString(upstreamRef),
			infoSign + "Branch name: " + color.CyanString(branchName),
		}
		print(boxify(strings.Join(lines, "\n")))

		var shouldContinue bool
		askOne(&survey.Confirm{Message: "Continue?"}, &shouldContinue)
		if !shouldContinue {
			return cli.Exit("", 1)
		}

		if context.Bool("skip-fetch") {
			color.White("-️ Skipping fetch")
		} else {
			color.Green("▶️ Fetching " + pullRemote)
			_, err = runCommand("git", "fetch", pullRemote)
			if err != nil {
				return err
			}
		}

		color.Green("▶️ Creating temporary directory")
		tmpDir, err := runCommand("mktemp", "-d")
		if err != nil {
			return err
		}

		color.Green("▶️ Creating temporary worktree in " + tmpDir)
		_, err = runCommand("git", "worktree", "add", "--no-track", "-b", branchName, tmpDir, upstreamRef)
		if err != nil {
			return err
		}
		cwd = tmpDir

		color.Green("▶️ cherry picking commits")
		args := append([]string{"cherry-pick"}, pickedCommitShas...)
		_, err = runCommand("git", args...)
		if err != nil {
			return err
		}

		color.Green("▶️ Pushing to " + pushRemote + "/" + branchName)
		_, err = runCommand("git", "push", "-u", pushRemote, branchName)
		if err != nil {
			return err
		}

		if !context.Bool("no-pr") {
			color.Green("▶️ Creating pull request")
			_, err = runCommand("gh", "pr", "create", "--fill", "--assignee", "@me", "--base", upstreamBranch)
			if err != nil {
				return err
			}
		}

		color.Green("▶️ Cleaning up temporary worktree")
		_, err = runCommand("git", "worktree", "remove", tmpDir)
		if err != nil {
			return err
		}

		return nil
	},
}

func boxify(message string) string {
	c := color.New(color.FgCyan)
	const (
		horizontal = "─"
		prefix     = "│  "
		suffix     = "  │"
	)
	lengthFn := func(s string) int {
		return utf8.RuneCountInString(stripansi.Strip(s))
	}
	parts := strings.Split(message, "\n")
	width := longestString(parts, lengthFn)
	horizontalLine := strings.Repeat(horizontal, width+lengthFn(prefix)+lengthFn(suffix)-2)
	out := c.Sprint("┌" + horizontalLine + "┐\n")
	for _, part := range parts {
		fill := strings.Repeat(" ", width-lengthFn(part))
		out += c.Sprint(prefix) + part + fill + c.Sprint(suffix) + "\n"
	}
	out += c.Sprint("└" + horizontalLine + "┘\n")
	return out
}

func longestString(options []string, lengthFn func(string) int) int {
	best := ""
	for _, option := range options {
		if lengthFn(option) > lengthFn(best) {
			best = option
		}
	}
	return lengthFn(best)
}

func askOne(p survey.Prompt, response interface{}, opts ...survey.AskOpt) {
	err := survey.AskOne(p, response, opts...)
	if err != nil {
		if err == terminal.InterruptErr {
			os.Exit(1)
		}
		log.Fatal(err)
	}
}

func getCommits(revisionRange string) ([]string, error) {
	recentCommits, err := runCommand("git", "show", "--quiet", "--pretty=format:%h %s", revisionRange)
	if err != nil {
		return nil, cli.Exit("Failed to retrieve recent commits", 1)
	}
	return strings.Split(recentCommits, "\n"), nil
}

func verifyCommits(commitShas []string) ([]string, error) {
	var verifiedShas []string
	for _, commitSha := range commitShas {
		fullCommitSha, err := runCommand("git", "rev-parse", "--quiet", "--verify", commitSha+"^{commit}")
		if err != nil {
			return nil, cli.Exit("Given commits are invalid", 1)
		}
		verifiedShas = append(verifiedShas, fullCommitSha)
	}
	return verifiedShas, nil
}

func runCommand(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	cmd.Dir = cwd
	println(cmd.String())
	output, err := cmd.Output()
	println(string(output))
	printCommandError(err)
	return strings.TrimSpace(string(output)), err
}

func printCommandError(err error) {
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			println(string(exiterr.Stderr))
		}
	}
}
