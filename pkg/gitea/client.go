package gitea

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

type Repo struct {
	Name        string
	CloneURL    string
	Description string
	Private     bool
	Fork        bool
}

type Client struct {
	url   string
	api   string
	token string
	user  string
	cl    *resty.Client
}

func New(url, token, user string) *Client {
	url = strings.TrimSuffix(url, "/")
	return &Client{
		url:   url,
		api:   url + "/api/v1",
		token: token,
		user:  user,
		cl:    resty.New(),
	}
}

func (c *Client) GetRepos() ([]Repo, error) {
	var rs []Repo
	pg := 1

	fmt.Printf("Fetching repos for %s...\n", c.user)

	for {
		var result []map[string]interface{}
		_, err := c.cl.R().SetResult(&result).Get(fmt.Sprintf("%s/user/repos?page=%d&per_page=100", c.api, pg))
		if err != nil {
			return rs, err
		}

		if len(result) == 0 {
			break
		}

		for _, r := range result {
			cloneURL, _ := r["clone_url"].(string)
			desc, _ := r["description"].(string)
			rs = append(rs, Repo{
				Name:        r["name"].(string),
				CloneURL:    cloneURL,
				Description: desc,
				Private:     r["private"].(bool),
				Fork:        r["fork"].(bool),
			})
		}

		fmt.Printf("  Page %d: %d repos\n", pg, len(result))
		pg++
	}

	fmt.Printf("Total: %d repos\n", len(rs))
	return rs, nil
}

func (c *Client) RepoExists(name string) bool {
	if c.token == "" || c.user == "" {
		return false
	}

	resp, _ := c.cl.R().Get(fmt.Sprintf("%s/repos/%s/%s", c.api, c.user, name))
	return resp.StatusCode() == 200
}

func (c *Client) CreateRepo(name, desc string, priv bool) bool {
	if c.token == "" {
		return false
	}

	if len(desc) > 255 {
		desc = desc[:255]
	}

	data := map[string]interface{}{
		"name":           name,
		"description":    desc,
		"private":        priv,
		"auto_init":      false,
		"default_branch": "main",
	}

	resp, err := c.cl.R().SetBody(data).Post(fmt.Sprintf("%s/user/repos", c.api))
	if err != nil {
		return false
	}

	if resp.StatusCode() == 201 {
		fmt.Printf("  Created repo %s\n", name)
		return true
	}
	if resp.StatusCode() == 409 {
		fmt.Printf("  Repo %s exists\n", name)
		return true
	}
	return false
}

func (c *Client) PushURL(name string) string {
	if c.token != "" && c.user != "" {
		return fmt.Sprintf("%s://%s:%s@%s/%s/%s.git", "https", c.user, c.token, strings.TrimPrefix(c.url, "https://"), c.user, name)
	}
	return fmt.Sprintf("%s/%s/%s.git", c.url, c.user, name)
}
