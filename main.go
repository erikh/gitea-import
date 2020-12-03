package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

var whitespaceRegexp = regexp.MustCompile(`\s+`)

func usage() {
	fmt.Println("usage: [url] [api-token] [change list filename]")
	fmt.Println("See README for more details.")
	fmt.Println("Author: Erik Hollensbe <github@hollensbe.org>")
	os.Exit(1)
}

func errOut(e error) {
	fmt.Fprintln(os.Stderr, e)
	os.Exit(2)
}

func main() {
	if len(os.Args) != 4 {
		usage()
	}

	url := os.Args[1]
	apiToken := os.Args[2]
	fileName := os.Args[3]

	client, err := gitea.NewClient(url, gitea.SetToken(apiToken))
	if err != nil {
		errOut(err)
	}

	u, _, err := client.GetMyUserInfo()
	if err != nil {
		errOut(err)
	}

	log.Println("Authenticated as:", u.UserName)
	log.Println("Reading change list")

	f, err := os.Open(fileName)
	if err != nil {
		errOut(err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)

	for s.Scan() {
		parts := whitespaceRegexp.Split(s.Text(), 2)
		if len(parts) != 2 {
			errOut(errors.New("invalid format for change list, please see README"))
		}

		fullName := parts[0]
		private, err := strconv.ParseBool(parts[1])
		if err != nil {
			errOut(fmt.Errorf("Invalid boolean value for repo %q: %w", fullName, err))
		}

		parts = strings.SplitN(fullName, "/", 2)

		if len(parts) != 2 {
			errOut(errors.New("invalid repository format"))
		}

		org := parts[0]
		repo := parts[1]

		log.Printf("PROCESSING: [org: %q] [repo: %q] [private: %v]", org, repo, private)
		gitPath := fullName + ".git"

		gitRepo, err := git.PlainOpen(gitPath)
		if err != nil {
			errOut(fmt.Errorf("While opening repository %q: %w", fullName, err))
		}

		ref, err := gitRepo.Head()
		if err != nil {
			errOut(err)
		}

		log.Println("Found default branch:", ref.Name().Short())

		// just try to create it; don't care about errors here.
		client.CreateOrg(gitea.CreateOrgOption{Name: org})

		if _, _, err := client.GetRepo(org, repo); err == nil {
			log.Printf("Repository '%s/%s' already exists; skipping", org, repo)
			continue
		}

		var giteaRepo *gitea.Repository
		options := gitea.CreateRepoOption{
			Name:          repo,
			Private:       private,
			DefaultBranch: ref.Name().Short(),
		}

		if org == u.UserName {
			giteaRepo, _, err = client.CreateRepo(options)
		} else {
			giteaRepo, _, err = client.CreateOrgRepo(org, options)
		}

		gitRepo.DeleteRemote("gitea-import")
		_, err = gitRepo.CreateRemote(&config.RemoteConfig{
			Name: "gitea-import",
			URLs: []string{giteaRepo.CloneURL},
		})
		if err != nil {
			errOut(fmt.Errorf("while creating remote for %q: %w", fullName, err))
		}

		err = gitRepo.Push(&git.PushOptions{
			RemoteName: "gitea-import",
			Auth: &http.BasicAuth{
				Username: u.UserName,
				Password: apiToken,
			},
		})
		if err != nil {
			errOut(fmt.Errorf("while pushing %q: %w", fullName, err))
		}
	}
}
