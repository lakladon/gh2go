package migrator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Repo struct {
	Name        string
	CloneURL    string
	Description string
	Private     bool
	Fork        bool
}

type RepoFetcher interface {
	GetRepos() ([]Repo, error)
}

type RepoCreator interface {
	RepoExists(name string) bool
	CreateRepo(name, desc string, priv bool) bool
	PushURL(name string) string
}

type Stats struct {
	Total       int
	Success     int
	Failed      int
	FailedRepos []string
}

type Migrator struct {
	src RepoFetcher
	dst RepoCreator
	st  Stats
	in  string
}

func New(src, dst interface{}, in string) *Migrator {
	return &Migrator{
		src: src.(RepoFetcher),
		dst: dst.(RepoCreator),
		in:  in,
	}
}

func (m *Migrator) Run(forks bool, retries int) error {
	repos, err := m.src.GetRepos()
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		fmt.Println("No repos found")
		return nil
	}

	m.st.Total = len(repos)

	if !forks {
		var f []Repo
		for _, r := range repos {
			if !r.Fork {
				f = append(f, r)
			}
		}
		repos = f
	}

	for i, r := range repos {
		fmt.Printf("\n%s\n[%d/%d] Repo: %s\n", strings.Repeat("=", 50), i+1, len(repos), r.Name)

		for att := 0; att < retries; att++ {
			if att > 0 {
				fmt.Printf("  Retry %d...\n", att+1)
				time.Sleep(2 * time.Second)
			}
			m.migrate(r)
			if !contains(m.st.FailedRepos, r.Name) {
				break
			}
		}
	}
	return nil
}

func (m *Migrator) migrate(r Repo) {
	td, err := os.MkdirTemp("", fmt.Sprintf("mig_%s_", r.Name))
	if err != nil {
		m.st.Failed++
		m.st.FailedRepos = append(m.st.FailedRepos, r.Name)
		return
	}
	defer os.RemoveAll(td)

	rp := filepath.Join(td, r.Name)
	os.MkdirAll(rp, 0755)

	if m.dst != nil {
		if !m.dst.RepoExists(r.Name) {
			m.dst.CreateRepo(r.Name, r.Description, r.Private)
		}
	}

	fmt.Printf("  Cloning %s...\n", r.CloneURL)
	if err := m.clone(r.CloneURL, rp); err != nil {
		fmt.Printf("  Clone failed: %v, trying mirror...\n", err)
		if err := m.mirror(r.CloneURL, rp); err != nil {
			m.st.Failed++
			m.st.FailedRepos = append(m.st.FailedRepos, r.Name)
			return
		}
	}

	if m.dst != nil {
		purl := m.dst.PushURL(r.Name)
		if err := m.push(rp, purl); err != nil {
			m.st.Failed++
			m.st.FailedRepos = append(m.st.FailedRepos, r.Name)
			return
		}
	}

	m.st.Success++
	fmt.Printf("  %s migrated OK\n", r.Name)
}

func (m *Migrator) clone(url, path string) error {
	return exec.Command("git", "clone", "--no-hardlinks", url, path).Run()
}

func (m *Migrator) mirror(url, path string) error {
	return exec.Command("git", "clone", "--mirror", url, path).Run()
}

func (m *Migrator) push(path, url string) error {
	if err := exec.Command("git", "remote", "set-url", "origin", url).Run(); err != nil {
		exec.Command("git", "remote", "add", "origin", url).Run()
	}

	if err := exec.Command("git", "push", "--all", "origin").Run(); err != nil {
		br, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
		exec.Command("git", "push", "-u", "origin", strings.TrimSpace(string(br))).Run()
	}

	exec.Command("git", "push", "--tags", "origin").Run()
	fmt.Printf("  Pushed to Gitea\n")
	return nil
}

func (m *Migrator) PrintSummary() {
	fmt.Printf("\n%s\nSUMMARY:\n%s\n", strings.Repeat("=", 50), strings.Repeat("=", 50))
	fmt.Printf("Total: %d, Success: %d, Failed: %d\n", m.st.Total, m.st.Success, m.st.Failed)
	if len(m.st.FailedRepos) > 0 {
		fmt.Println("Failed repos:")
		for _, r := range m.st.FailedRepos {
			fmt.Printf("  %s\n", r)
		}
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
