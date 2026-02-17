package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v45/github"
)

type Repo struct {
	Name        string
	CloneURL    string
	Description string
	Private     bool
	Fork        bool
}

type Client struct {
	un string
	cl *github.Client
}

func New(un string) *Client {
	return &Client{
		un: un,
		cl: github.NewClient(nil),
	}
}

func (c *Client) GetRepos() ([]Repo, error) {
	var rs []Repo
	pg := 1

	fmt.Printf("Fetching repos for %s...\n", c.un)

	for {
		repos, resp, err := c.cl.Repositories.List(context.Background(), c.un, &github.RepositoryListOptions{
			ListOptions: github.ListOptions{Page: pg, PerPage: 100},
			Type:        "all",
			Sort:        "updated",
		})
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			return rs, err
		}

		for _, r := range repos {
			rs = append(rs, Repo{
				Name:        r.GetName(),
				CloneURL:    r.GetCloneURL(),
				Description: r.GetDescription(),
				Private:     r.GetPrivate(),
				Fork:        r.GetFork(),
			})
		}

		fmt.Printf("  Page %d: %d repos\n", pg, len(repos))

		if resp.NextPage == 0 || len(repos) == 0 {
			break
		}
		pg = resp.NextPage
	}

	fmt.Printf("Total: %d repos\n", len(rs))
	return rs, nil
}
