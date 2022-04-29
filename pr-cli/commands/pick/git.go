package pick

import (
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"pr-cli/lib/shell"
	"strings"
)

type gitPickSettings struct {
	doFetch             bool
	doCreatePullRequest bool

	pullRemote string
	pushRemote string

	branchName     string
	upstreamBranch string

	commits []*commit
}

func cherryPickToNewBranch(settings *gitPickSettings) error {
	if !settings.doFetch {
		color.White("-️ Skipping fetch")
	} else {
		color.Green("▶️ Fetching " + settings.pullRemote)
		_, err := shell.RunCommand("git", "fetch", settings.pullRemote)
		if err != nil {
			return err
		}
	}

	color.Green("▶️ Creating temporary directory")
	tmpDir, err := shell.RunCommand("mktemp", "-d")
	if err != nil {
		return err
	}

	color.Green("▶️ Creating temporary worktree in " + tmpDir)
	upstreamRef := settings.pullRemote + "/" + settings.upstreamBranch
	_, err = shell.RunCommand("git", "worktree", "add", "--no-track", "-b", settings.branchName, tmpDir, upstreamRef)
	if err != nil {
		return err
	}

	shell.SetCwd(tmpDir)

	color.Green("▶️ cherry picking commits")
	var pickedCommitSHAs = make([]string, len(settings.commits))
	for i, commit := range settings.commits {
		pickedCommitSHAs[i] = commit.sha
	}
	args := append([]string{"cherry-pick"}, pickedCommitSHAs...)
	_, err = shell.RunCommand("git", args...)
	if err != nil {
		return err
	}

	color.Green("▶️ Pushing to " + settings.pushRemote + "/" + settings.branchName)
	_, err = shell.RunCommand("git", "push", "-u", settings.pushRemote, settings.branchName)
	if err != nil {
		return err
	}

	if settings.doCreatePullRequest {
		color.Green("▶️ Creating pull request")
		_, err = shell.RunCommand("gh", "pr", "create", "--fill", "--assignee", "@me", "--base", settings.upstreamBranch)
		if err != nil {
			return err
		}
	}

	color.Green("▶️ Cleaning up temporary worktree")
	_, err = shell.RunCommand("git", "worktree", "remove", tmpDir)
	if err != nil {
		return err
	}

	return nil
}

type commit struct {
	sha     string
	message string
}

func getCommits(revisionRange string) ([]string, error) {
	gitShowOutput, err := shell.RunCommandQuiet("git", "show", "--quiet", "--pretty=format:%h %s", revisionRange)
	if err != nil {
		return nil, cli.Exit("Failed to retrieve recent commits", 1)
	}
	if len(gitShowOutput) < 1 {
		return []string{}, nil
	}
	return strings.Split(gitShowOutput, "\n"), nil
}

func convertLinesToCommits(lines []string) []*commit {
	commits := make([]*commit, len(lines))
	for i, line := range lines {
		commits[i] = convertLineToCommit(line)
	}
	return commits
}

func convertLineToCommit(line string) *commit {
	parts := strings.SplitN(line, " ", 2)
	return &commit{
		sha:     parts[0],
		message: parts[1],
	}
}

func verifyCommits(commitSHAs []string) ([]string, error) {
	verifiedSHAs := make([]string, len(commitSHAs))
	for i, commitSha := range commitSHAs {
		fullCommitSha, err := shell.RunCommand("git", "rev-parse", "--quiet", "--verify", commitSha+"^{commit}")
		if err != nil {
			return nil, cli.Exit("Given commits are invalid", 1)
		}
		verifiedSHAs[i] = fullCommitSha
	}
	return verifiedSHAs, nil
}
