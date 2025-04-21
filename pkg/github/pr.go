package github

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Title       string
	Description string
	Branch      string
	BaseBranch  string
	Files       map[string]string // Filename -> Content
}

// GitHubClient handles GitHub API operations
type GitHubClient struct {
	client      *github.Client
	ctx         context.Context
	Token       string
	Owner       string
	Repo        string
	CommitName  string
	CommitEmail string
}

// NewGitHubClient creates a new GitHub client
func NewGitHubClient(token, owner, repo string) (*GitHubClient, error) {
	if token == "" {
		return nil, errors.New("GitHub token is required")
	}

	// Get commit identity from git config or use defaults
	commitName := "Dai Dependency Bot"
	commitEmail := "dai-bot@users.noreply.github.com"

	// Try to get user's git config if available
	nameCmd, _ := runGitCommand("config", "user.name")
	emailCmd, _ := runGitCommand("config", "user.email")

	if nameCmd != "" {
		commitName = strings.TrimSpace(nameCmd)
	}

	if emailCmd != "" {
		commitEmail = strings.TrimSpace(emailCmd)
	}

	// Create an OAuth2 client with the token
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	// Create the GitHub client
	client := github.NewClient(tc)

	return &GitHubClient{
		client:      client,
		ctx:         ctx,
		Token:       token,
		Owner:       owner,
		Repo:        repo,
		CommitName:  commitName,
		CommitEmail: commitEmail,
	}, nil
}

// runGitCommand runs a git command and returns its output
func runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		// Check if it's an ExitError, which contains stderr output
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git command failed: %s: %w", string(exitErr.Stderr), err)
		}
		return "", fmt.Errorf("git command failed: %w", err)
	}

	return string(output), nil
}

// GetRepoDetails extracts owner and repo from the current git remote
func GetRepoDetails() (string, string, error) {
	// Try to get the remote URL
	remoteURL, err := runGitCommand("remote", "get-url", "origin")
	if err != nil {
		return "", "", fmt.Errorf("failed to get git remote URL: %w", err)
	}

	remoteURL = strings.TrimSpace(remoteURL)

	// Parse the GitHub URL to extract owner and repo
	owner, repo := "", ""

	// Handle SSH URLs like: git@github.com:owner/repo.git
	if strings.HasPrefix(remoteURL, "git@github.com:") {
		parts := strings.SplitN(strings.TrimPrefix(remoteURL, "git@github.com:"), "/", 2)
		if len(parts) == 2 {
			owner = parts[0]
			repo = strings.TrimSuffix(parts[1], ".git")
		}
	}

	// Handle HTTPS URLs like: https://github.com/owner/repo.git
	if strings.HasPrefix(remoteURL, "https://github.com/") {
		parts := strings.SplitN(strings.TrimPrefix(remoteURL, "https://github.com/"), "/", 2)
		if len(parts) == 2 {
			owner = parts[0]
			repo = strings.TrimSuffix(parts[1], ".git")
		}
	}

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("could not parse GitHub owner and repo from URL: %s", remoteURL)
	}

	return owner, repo, nil
}

// CreatePullRequest creates a new pull request on GitHub
func (c *GitHubClient) CreatePullRequest(pr *PullRequest) (string, error) {
	// Default to main if no base branch is provided
	if pr.BaseBranch == "" {
		pr.BaseBranch = "main"
	}

	// First, create a new branch if it doesn't exist
	if err := c.CreateBranch(pr.Branch, pr.BaseBranch); err != nil {
		return "", fmt.Errorf("failed to create branch: %w", err)
	}

	// Commit the files to the branch
	if err := c.CommitFiles(pr.Branch, pr.Files, "Update dependencies"); err != nil {
		return "", fmt.Errorf("failed to commit files: %w", err)
	}

	// Create the pull request
	newPR := &github.NewPullRequest{
		Title:               github.String(pr.Title),
		Head:                github.String(pr.Branch),
		Base:                github.String(pr.BaseBranch),
		Body:                github.String(pr.Description),
		MaintainerCanModify: github.Bool(true),
		Draft:               github.Bool(true), // Start as draft PR
	}

	createdPR, _, err := c.client.PullRequests.Create(c.ctx, c.Owner, c.Repo, newPR)
	if err != nil {
		return "", fmt.Errorf("failed to create pull request: %w", err)
	}

	return createdPR.GetHTMLURL(), nil
}

// CreateBranch creates a new branch
func (c *GitHubClient) CreateBranch(name, base string) error {
	// Get the reference for the base branch
	baseRef, _, err := c.client.Git.GetRef(c.ctx, c.Owner, c.Repo, "refs/heads/"+base)
	if err != nil {
		return fmt.Errorf("failed to get base branch reference: %w", err)
	}

	// Create new reference (branch)
	_, _, err = c.client.Git.CreateRef(c.ctx, c.Owner, c.Repo, &github.Reference{
		Ref:    github.String("refs/heads/" + name),
		Object: baseRef.Object,
	})

	// If the branch already exists, that's fine
	if err != nil && !strings.Contains(err.Error(), "Reference already exists") {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

// CommitFiles commits files to a branch
func (c *GitHubClient) CommitFiles(branch string, files map[string]string, message string) error {
	// Get the latest commit on the branch to get its tree
	ref, _, err := c.client.Git.GetRef(c.ctx, c.Owner, c.Repo, "refs/heads/"+branch)
	if err != nil {
		return fmt.Errorf("failed to get branch reference: %w", err)
	}

	// Get the commit object that the branch reference points to
	commit, _, err := c.client.Git.GetCommit(c.ctx, c.Owner, c.Repo, ref.Object.GetSHA())
	if err != nil {
		return fmt.Errorf("failed to get commit: %w", err)
	}

	// Get the tree that the commit points to
	baseTreeSHA := commit.Tree.GetSHA()

	// Create blobs for each file
	var entries []*github.TreeEntry
	for path, content := range files {
		// Create a blob for the file
		blob := &github.Blob{
			Content:  github.String(content),
			Encoding: github.String("utf-8"),
		}

		createdBlob, _, err := c.client.Git.CreateBlob(c.ctx, c.Owner, c.Repo, blob)
		if err != nil {
			return fmt.Errorf("failed to create blob for %s: %w", path, err)
		}

		// Add this blob as a tree entry
		entries = append(entries, &github.TreeEntry{
			Path: github.String(path),
			Mode: github.String("100644"), // Regular file
			Type: github.String("blob"),
			SHA:  createdBlob.SHA,
		})
	}

	// Create a tree with those blobs
	newTree, _, err := c.client.Git.CreateTree(c.ctx, c.Owner, c.Repo, baseTreeSHA, entries)
	if err != nil {
		return fmt.Errorf("failed to create tree: %w", err)
	}

	// Create a commit with this tree
	newCommit := &github.Commit{
		Message: github.String(message),
		Tree:    newTree,
		Parents: []*github.Commit{{SHA: commit.SHA}},
		Author: &github.CommitAuthor{
			Name:  github.String(c.CommitName),
			Email: github.String(c.CommitEmail),
			Date:  &github.Timestamp{Time: time.Now()},
		},
	}

	createdCommit, _, err := c.client.Git.CreateCommit(c.ctx, c.Owner, c.Repo, newCommit, nil)
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	// Update the branch reference to point to the new commit
	ref.Object.SHA = createdCommit.SHA
	_, _, err = c.client.Git.UpdateRef(c.ctx, c.Owner, c.Repo, ref, false)
	if err != nil {
		return fmt.Errorf("failed to update branch reference: %w", err)
	}

	return nil
}

// ListPullRequests lists open pull requests
func (c *GitHubClient) ListPullRequests() ([]*github.PullRequest, error) {
	// List open pull requests
	prs, _, err := c.client.PullRequests.List(c.ctx, c.Owner, c.Repo, &github.PullRequestListOptions{
		State: "open",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	return prs, nil
}

// GetDefaultBranch gets the default branch of the repository
func (c *GitHubClient) GetDefaultBranch() (string, error) {
	repo, _, err := c.client.Repositories.Get(c.ctx, c.Owner, c.Repo)
	if err != nil {
		return "main", fmt.Errorf("failed to get repository info: %w", err)
	}

	return repo.GetDefaultBranch(), nil
}
