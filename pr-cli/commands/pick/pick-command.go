package pick

import "C"
import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"pr-cli/lib/shell"
	"pr-cli/lib/slug"
	"strings"
)

var Cmd = cli.Command{
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
			return cli.Exit(fmt.Sprintf("no %s branch provided", pullRemote), 1)
		}
		upstreamRef := pullRemote + "/" + upstreamBranch

		pickedCommits, err := parseCommits(context, upstreamRef)
		if err != nil {
			return err
		}
		if len(pickedCommits) < 1 {
			return cli.Exit("No commits chosen", 1)
		}

		branchName, err := getBranchName(context, pickedCommits)
		if err != nil {
			return err
		}

		settings := &gitPickSettings{
			doFetch:             !context.Bool("skip-fetch"),
			doCreatePullRequest: !context.Bool("no-pr"),
			pullRemote:          pullRemote,
			pushRemote:          pushRemote,
			branchName:          branchName,
			upstreamBranch:      upstreamBranch,
			commits:             pickedCommits,
		}

		if !confirmSettings(settings) {
			return cli.Exit("", 1)
		}

		err = cherryPickToNewBranch(settings)
		if err != nil {
			return err
		}

		return nil
	},
}

func getBranchName(context *cli.Context, pickedCommits []*commit) (string, error) {
	var branchName string
	branchValidator := survey.MinLength(3)
	if context.IsSet("branch") {
		branchName = context.String("branch")
	} else {
		var suggestedBranchName string
		if len(pickedCommits) == 1 {
			suggestedBranchName = generateBranchNameFromCommitMessage(pickedCommits[0].message)
		}

		shell.AskOne(&survey.Input{
			Message: "Branch name:",
			Default: suggestedBranchName,
		}, &branchName, survey.WithValidator(branchValidator))
	}
	err := branchValidator(branchName)
	if err != nil {
		return "", err
	}
	return branchName, nil
}

func parseCommits(context *cli.Context, upstreamRef string) ([]*commit, error) {
	if context.IsSet("commits") {
		return getCommitsFromSHAs(context.StringSlice("commits"))
	} else {
		return getCommitsByAsking(upstreamRef)
	}
}

func getCommitsFromSHAs(commitSHAs []string) ([]*commit, error) {
	providedCommitSHAs, err := verifyCommits(commitSHAs)
	if err != nil {
		return nil, err
	}

	pickedCommitLines, err := getCommits(strings.Join(providedCommitSHAs, " "))
	if err != nil {
		return nil, err
	}

	return convertLinesToCommits(pickedCommitLines), nil
}

func getCommitsByAsking(upstreamRef string) ([]*commit, error) {
	recentCommitLines, err := getCommits(upstreamRef + "..")
	if err != nil {
		return nil, err
	}

	if len(recentCommitLines) < 1 {
		return nil, cli.Exit("No commits to pick", 1)
	}

	var pickedCommitLines []string

	shell.AskOne(&survey.MultiSelect{
		Message: "Which commits do you want to pick?",
		Options: recentCommitLines,
	}, &pickedCommitLines, survey.WithKeepFilter(true), survey.WithValidator(survey.MinItems(1)))

	return convertLinesToCommits(pickedCommitLines), nil
}

func generateBranchNameFromCommitMessage(message string) string {
	return slug.Slugify(message)
}

func confirmSettings(settings *gitPickSettings) (shouldContinue bool) {
	infoSign := color.HiGreenString("ℹ ")
	var commitLines string
	for _, commit := range settings.commits {
		commitLines += "\n" + "   └▷ " + color.MagentaString(commit.sha) + " " + color.CyanString(commit.message)
	}
	lines := []string{
		infoSign + "About to cherry pick commits:" + commitLines,
		infoSign + "Base branch: " + color.CyanString(settings.pullRemote) + "/" + color.CyanString(settings.upstreamBranch),
		infoSign + "Branch name: " + color.CyanString(settings.branchName),
	}
	println(shell.FormatInBox(strings.Join(lines, "\n")))

	shell.AskOne(&survey.Confirm{Message: "Continue?"}, &shouldContinue)
	return shouldContinue
}
