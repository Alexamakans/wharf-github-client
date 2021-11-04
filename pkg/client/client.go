package client

import (
	"fmt"
	"strconv"

	"github.com/Alexamakans/wharf-common-api-client/pkg/remoteprovider"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Client implements remoteprovider.Client.
type Client struct {
	remoteprovider.BaseClient
}

func (c *Client) FetchFile(projectIdentifier remoteprovider.ProjectIdentifier, fileName string) ([]byte, error) {
	return []byte{}, nil
}

func (c *Client) FetchBranches(projectIdentifier remoteprovider.ProjectIdentifier) ([]remoteprovider.WharfBranch, error) {
	return []remoteprovider.WharfBranch{}, nil
}

func (c *Client) FetchProjectByGroupAndProjectName(groupName, projectName string) (remoteprovider.WharfProject, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.Token})
	tc := oauth2.NewClient(c.Context, ts)
	githubClient, err := github.NewEnterpriseClient(c.RemoteProviderURL, "", tc)
	if err != nil {
		return remoteprovider.WharfProject{}, fmt.Errorf("failed connecting to github using remote provider URL %q: %w", c.RemoteProviderURL, err)
	}

	githubRepo, _, err := githubClient.Repositories.Get(c.Context, groupName, projectName)
	if err != nil {
		return remoteprovider.WharfProject{}, fmt.Errorf("fetching github project '%s/%s' failed on %q: %w", groupName, projectName, c.RemoteProviderURL, err)
	}

	var project remoteprovider.WharfProject
	project.RemoteProjectID = strconv.FormatInt(githubRepo.GetID(), 10)
	project.GitURL = githubRepo.GetSSHURL()
	project.Name = githubRepo.GetName()
	project.GroupName = githubRepo.GetOwner().GetName()

	return project, nil
}

func (c *Client) WharfProjectToIdentifier(project remoteprovider.WharfProject) remoteprovider.ProjectIdentifier {
	return remoteprovider.ProjectIdentifier{
		Values: []string{project.RemoteProjectID},
	}
}
