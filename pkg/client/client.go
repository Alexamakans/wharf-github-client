package client

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/Alexamakans/wharf-common-api-client/pkg/remoteprovider"
	"github.com/google/go-github/github"
	"github.com/iver-wharf/wharf-api/pkg/model/database"
	"golang.org/x/oauth2"
)

// Client implements remoteprovider.Client.
type Client struct {
	remoteprovider.BaseClient
}

func NewClient(ctx context.Context, token, remoteProviderURL string) *Client {
	return &Client{
		*remoteprovider.NewClient(ctx, token, remoteProviderURL),
	}
}

func (c *Client) FetchFile(projectIdentifier remoteprovider.ProjectIdentifier, fileName string) ([]byte, error) {
	githubClient, err := c.initGithubClient()
	if err != nil {
		return []byte{}, nil
	}

	repo, err := c.getRepo(githubClient, projectIdentifier.Values[0])
	if err != nil {
		return []byte{}, nil
	}

	fileContent, _, _, err := githubClient.Repositories.GetContents(c.Context, repo.GetOwner().GetName(), repo.GetName(), fileName, nil)
	if err != nil {
		return []byte{}, err
	}

	return b64.StdEncoding.DecodeString(*fileContent.Content)
}

func (c *Client) FetchBranches(projectIdentifier remoteprovider.ProjectIdentifier) ([]remoteprovider.WharfBranch, error) {
	githubClient, err := c.initGithubClient()
	if err != nil {
		return []remoteprovider.WharfBranch{}, nil
	}

	repo, err := c.getRepo(githubClient, projectIdentifier.Values[0])
	if err != nil {
		return []remoteprovider.WharfBranch{}, nil
	}

	branches, _, err := githubClient.Repositories.ListBranches(c.Context, repo.GetOwner().GetName(), repo.GetName(), nil)
	if err != nil {
		return []remoteprovider.WharfBranch{}, err
	}

	wharfBranches := make([]remoteprovider.WharfBranch, 0, len(branches))
	for _, branch := range branches {
		wharfBranches = append(wharfBranches, remoteprovider.WharfBranch{
			Name:    branch.GetName(),
			Default: branch.GetName() == repo.GetDefaultBranch()})
		if err != nil {
			break
		}
	}

	return wharfBranches, nil
}

func (c *Client) FetchProjectByGroupAndProjectName(groupName, projectName string) (remoteprovider.WharfProject, error) {
	githubClient, err := c.initGithubClient()
	if err != nil {
		return remoteprovider.WharfProject{}, nil
	}

	repo, _, err := githubClient.Repositories.Get(c.Context, groupName, projectName)
	if err != nil {
		return remoteprovider.WharfProject{}, fmt.Errorf("fetching github project '%s/%s' failed on %q: %w", groupName, projectName, c.RemoteProviderURL, err)
	}

	project := remoteprovider.WharfProject{
		Project: database.Project{
			Name:        repo.GetName(),
			GroupName:   repo.GetOwner().GetLogin(),
			Description: repo.GetDescription(),
			AvatarURL:   *repo.GetOwner().AvatarURL,
			GitURL:      *repo.GitURL,
		},
		RemoteProjectID: strconv.FormatInt(repo.GetID(), 10)}

	return project, nil
}

func (c *Client) WharfProjectToIdentifier(project remoteprovider.WharfProject) remoteprovider.ProjectIdentifier {
	return remoteprovider.ProjectIdentifier{
		Values: []string{project.RemoteProjectID},
	}
}

func (c *Client) initGithubClient() (*github.Client, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.Token})
	tc := oauth2.NewClient(c.Context, ts)
	githubClient, err := github.NewEnterpriseClient(c.RemoteProviderURL, "", tc)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to github using remote provider URL %q: %w", c.RemoteProviderURL, err)
	}

	return githubClient, nil
}

func (c *Client) getRepo(githubClient *github.Client, remoteProjectID string) (*github.Repository, error) {
	id, err := strconv.ParseInt(remoteProjectID, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing remote project id failed %q: %w", remoteProjectID, err)
	}

	repo, _, err := githubClient.Repositories.GetByID(c.Context, id)
	if err != nil {
		return nil, fmt.Errorf("fetching github project with ID %q failed on %q: %w", remoteProjectID, c.RemoteProviderURL, err)
	}

	if repo.GetOwner().GetName() == "" {
		repo.Owner.Name = &strings.Split(repo.GetFullName(), "/")[0]
	}

	return repo, nil
}
