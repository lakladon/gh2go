package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/lakladon/gh2go/pkg/gitea"
	"github.com/lakladon/gh2go/pkg/github"
	"github.com/lakladon/gh2go/pkg/migrator"
)

func main() {
	ghUser := flag.String("github-user", "", "GitHub username")
	gtURL := flag.String("gitea-url", "", "Gitea URL")
	gtUser := flag.String("gitea-username", "", "Gitea username")
	gtToken := flag.String("gitea-token", "", "Gitea API token (required)")
	mForks := flag.Bool("migrate-forks", false, "Include forks")
	retries := flag.Int("retries", 2, "Retry count")

	flag.Parse()

	if *ghUser == "" {
		fmt.Println("Error: --github-user is required")
		os.Exit(1)
	}

	if *gtURL == "" {
		fmt.Println("Error: --gitea-url is required")
		os.Exit(1)
	}

	if *gtUser == "" {
		fmt.Println("Error: --gitea-username is required")
		os.Exit(1)
	}

	if *gtToken == "" {
		*gtToken = os.Getenv("GITEA_TOKEN")
	}

	if *gtToken == "" {
		fmt.Println("Error: --gitea-token is required")
		os.Exit(1)
	}

	if err := exec.Command("git", "--version").Run(); err != nil {
		fmt.Println("Error: git not found")
		os.Exit(1)
	}

	ghClient := github.New(*ghUser)
	gtClient := gitea.New(*gtURL, *gtToken, *gtUser)
	m := migrator.New(ghClient, gtClient, "github")

	m.Run(*mForks, *retries)
	m.PrintSummary()
}
