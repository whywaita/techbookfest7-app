package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gopkg.in/go-playground/webhooks.v5/github"
)

const (
	path = "/webhooks"
)

type Config struct {
	TerraformingUrl  string
	KustomizationUrl string
	GithubSecret     string
}

var config Config

func main() {
	err := setConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	http.HandleFunc(path, handleRequest)
	http.HandleFunc("/healthy", healthRequest)
	http.ListenAndServe(":80", nil)
}

func setConfig() error {
	config.TerraformingUrl = os.Getenv("TERRAFORMING_URL")
	config.KustomizationUrl = os.Getenv("KUSTOMIZATION_URL")
	config.GithubSecret = os.Getenv("GITHUB_SECRET")
	if config.TerraformingUrl == "" || config.KustomizationUrl == "" || config.GithubSecret == "" {
		return errors.New("TERRAFORMING_URL or KUSTOMIZATION_URL or GITHUB_SECRET are not set")
	}
	return nil
}

func healthRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var hook *github.Webhook
	if config.GithubSecret != "" {
		hook, _ = github.New(github.Options.Secret(config.GithubSecret))
	} else {
		hook, _ = github.New()
	}

	payload, err := hook.Parse(r, github.ReleaseEvent, github.PullRequestEvent)
	if err != nil {
		if err == github.ErrEventNotFound {
			fmt.Println(err)
		}
	}
	switch payload.(type) {
	case github.PullRequestPayload:
		pullRequest := payload.(github.PullRequestPayload)
		fmt.Printf("%+v\n", pullRequest.Action)
		switch pullRequest.Action {
		case "opened":
			err := nextCreateAction(pullRequest.PullRequest.ID, pullRequest.Action, pullRequest.PullRequest.Title, pullRequest.Sender.HTMLURL+"/"+pullRequest.Repository.Name, pullRequest.PullRequest.Head.Ref)
			if err != nil {
				fmt.Println(err)
			}
		case "closed":
			err, err2 := nextDeleteAction(pullRequest.PullRequest.ID, pullRequest.Action, pullRequest.PullRequest.Title, pullRequest.Sender.HTMLURL+"/"+pullRequest.Repository.Name, pullRequest.PullRequest.Head.Ref)
			if err != nil {
				fmt.Println(err, err2)
			}
		}
	}
}

func nextCreateAction(id int64, action, title, repoUrl, branchName string) error {
	if strings.Contains(title, "New Cluster") {
		return newClusterGenerate(id, repoUrl, branchName)
	}
	return newAppGenerate(id, repoUrl, branchName)
}

func nextDeleteAction(id int64, action, title, repoUrl, branchName string) (error, error) {
	var clusterErr, appErr error
	appErr = deleteApp(id, repoUrl, branchName)
	if strings.Contains(title, "New Cluster") {
		clusterErr = deleteCluster(id, repoUrl, branchName)
	}
	return clusterErr, appErr
}

func newClusterGenerate(id int64, repoUrl, branchName string) error {
	return httpPost(id, repoUrl, branchName, config.TerraformingUrl+"/apply", config.KustomizationUrl+"/apply")
}

func newAppGenerate(id int64, repoUrl, branchName string) error {
	return httpPost(id, repoUrl, branchName, config.KustomizationUrl+"/apply", "")
}

func deleteCluster(id int64, repoUrl, branchName string) error {
	return httpPost(id, repoUrl, branchName, config.TerraformingUrl+"/destroy", "")
}

func deleteApp(id int64, repoUrl, branchName string) error {
	return httpPost(id, repoUrl, branchName, config.KustomizationUrl+"/delete", "")
}

func httpPost(id int64, repoUrl, branchName, url, subUrl string) error {
	var jsonStr string
	if subUrl != "" {
		jsonStr = `{"id":"` + strconv.FormatInt(id, 10) + `","repoURL":"` + repoUrl + `","branchName":"` +
			branchName + `","kustomizationUrl":"` + subUrl + `"}`
	} else {
		jsonStr = `{"id":"` + strconv.FormatInt(id, 10) + `","repoURL":"` + repoUrl + `","branchName":"` + branchName + `"}`
	}
	fmt.Printf("Request: " + jsonStr + "\n")

	req, err := http.NewRequest(
		"POST",
		"http://"+url,
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf("statusCode: %d, Body: %+v\n", resp.StatusCode, resp.Body)
	}

	return err
}
