package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	pipeline "github.com/mattn/go-pipeline"
	git "gopkg.in/src-d/go-git.v4"
	. "gopkg.in/src-d/go-git.v4/_examples"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

const (
	applyPath  = "/apply"
	deletePath = "/delete"
	outputDir  = "./tmp"
)

type Config struct {
	BranchName    string
	RepoUrl       string
	GitUser       string
	PersonalToken string
	GCPProject    string
	GCPUser       string
}

type Message struct {
	Id         string `json:"id"`
	RepoUrl    string `json:"repoUrl"`
	BranchName string `json:"branchName"`
}

var envConfig Config

func main() {
	err := setConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	http.HandleFunc(applyPath, applyRequest)
	http.HandleFunc(deletePath, deleteRequest)
	http.ListenAndServe(":80", nil)

}

func setConfig() error {
	envConfig.GitUser = os.Getenv("GIT_USER")
	envConfig.PersonalToken = os.Getenv("PERSONAL_TOKEN")
	envConfig.GCPProject = os.Getenv("GCP_PROJECT")
	envConfig.GCPUser = os.Getenv("GCP_USER")
	if envConfig.GitUser == "" || envConfig.PersonalToken == "" || envConfig.GCPProject == "" || envConfig.GCPUser == "" {
		return errors.New("GIT_USER or PERSONAL_TOKEN or GCP_PROJECT or GCP_USER are not set")
	}
	return nil
}

func getClusterCredential(id, region string) (string, error) {
	fmt.Println("run cp /root/.config/gcloud/credentials.db.bak /root/.config/gcloud/credentials.db")
	out, err := pipeline.Output(
		[]string{"cp", "/root/.config/gcloud/credentials.db.bak", "/root/.config/gcloud/credentials.db"},
	)

	if err != nil {
		return "`cp` is failed", err
	}

	fmt.Println("run gcloud config set account " + envConfig.GCPUser)
	out, err = pipeline.Output(
		[]string{"gcloud", "config", "set", "account", envConfig.GCPUser},
	)

	if err != nil {
		return "`gcloud config set account` is failed", err
	}

	fmt.Println(string(out))

	fmt.Println("run gcloud config set project " + envConfig.GCPProject)
	out, err = pipeline.Output(
		[]string{"gcloud", "config", "set", "project", envConfig.GCPProject},
	)

	if err != nil {
		return "`gcloud config set project` is failed", err
	}

	fmt.Println(string(out))

	fmt.Println("run gcloud container clusters get-credentials dev-" + id + "-sample-app " + region)
	out, err = pipeline.Output(
		[]string{"gcloud", "container", "clusters", "get-credentials", "dev-" + id + "-sample-app", region},
	)

	fmt.Println(string(out))

	if err == nil {
		return "", nil
	}

	fmt.Println("run gcloud container clusters get-credentials dev-sample-app " + region)
	out, err = pipeline.Output(
		[]string{"gcloud", "container", "clusters", "get-credentials", "dev-sample-app", region},
	)

	if err != nil {
		return "`gcloud container cluster get-credential` is failed", err
	}

	fmt.Println(string(out))

	return "", nil
}

func isError(message string, err error, w http.ResponseWriter, code int) bool {
	if err != nil {
		fmt.Println(message, " error : ", err)
		http.Error(w, err.Error(), code)
		return true
	}
	return false
}

func applyRequest(w http.ResponseWriter, r *http.Request) {
	// sample '{"id":1234,"repoUrl":"https://github.com/username/reponame","branchName":"feature/hogehoge", "clusterName":"dev-1234"}'
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if isError("", err, w, 500) {
		return
	}

	// Unmarshal
	var msg Message
	err = json.Unmarshal(b, &msg)
	if isError("", err, w, 500) {
		return
	}

	errMessage, err := checkout(msg.RepoUrl, msg.BranchName, outputDir, envConfig.GitUser, envConfig.PersonalToken)
	if isError(errMessage, err, w, 500) {
		return
	}

	errMessage, err = generateVariables(outputDir+"/sample-app/kustomize"+"/overlays/dev", msg.Id)
	if isError(errMessage, err, w, 500) {
		return
	}

	errMessage, err = getClusterCredential(msg.Id, "--region=asia-northeast1")
	if isError(errMessage, err, w, 500) {
		return
	}

	errMessage, err = Apply(outputDir + "/sample-app/kustomize" + "/overlays/dev")
	if isError(errMessage, err, w, 500) {
		return
	}

	output, err := json.Marshal(msg)
	if isError("", err, w, 500) {
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
}

func deleteRequest(w http.ResponseWriter, r *http.Request) {
	// '{"id":1234,"repoUrl":"https://github.com/username/reponame","branchName":"feature/hogehoge"}'
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if isError("", err, w, 500) {
		return
	}

	// Unmarshal
	var msg Message
	err = json.Unmarshal(b, &msg)
	if isError("", err, w, 500) {
		return
	}

	errMessage, err := checkout(msg.RepoUrl, msg.BranchName, outputDir, envConfig.GitUser, envConfig.PersonalToken)
	if isError(errMessage, err, w, 500) {
		return
	}

	errMessage, err = generateVariables(outputDir+"/sample-app/kustomize"+"/overlays/dev", msg.Id)
	if isError(errMessage, err, w, 500) {
		return
	}

	errMessage, err = getClusterCredential(msg.Id, "--region=asia-northeast1")
	if isError(errMessage, err, w, 500) {
		return
	}

	errMessage, err = Delete(outputDir + "/sample-app/kustomize" + "/overlays/dev")
	if isError(errMessage, err, w, 500) {
		return
	}

	output, err := json.Marshal(msg)
	if isError("", err, w, 500) {
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
}

func checkout(url, branch, directory, user, token string) (string, error) {
	var err error
	Info("git clone %s %s", url, directory)

	r, err := git.PlainClone(directory, false, &git.CloneOptions{
		Auth: &githttp.BasicAuth{
			Username: user,
			Password: token,
		},
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		return "Clone is failed", err
	}

	w, err := r.Worktree()
	if err != nil {
		return "Can't get Worktree", err
	}

	err = r.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
		Auth: &githttp.BasicAuth{
			Username: user,
			Password: token,
		},
	})
	if err != nil {
		return "Fetch branch is failed", err
	}

	branchRef := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch))
	err = w.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
		Force:  true,
	})
	if err != nil {
		return "Checkout branch is failed", err
	}

	// ... retrieving the branch being pointed by HEAD
	ref, err := r.Head()
	if err != nil {
		return "Can't get Head", err
	}
	fmt.Println(ref.Hash())

	// ... retrieving the commit object
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return "Can't get CommitObject", err
	}

	fmt.Println(commit)

	return "", nil
}

func generateVariables(dir, id string) (string, error) {
	fmt.Println("change " + dir + "/kustomizaiton.yaml")

	_, err := pipeline.Output(
		[]string{"sed", "-i", "-e", "s/dev-/dev-" + id + "-/g", dir + "/kustomization.yaml"},
	)
	if err != nil {
		return "Change environment name is failed in kustomization.yaml", err
	}

	return "", nil
}

func Apply(dir string) (string, error) {
	defer removeContents(outputDir)
	fmt.Println("run kustomize build")

	out, err := pipeline.Output(
		[]string{"kustomize", "build", dir},
		[]string{"kubectl", "apply", "-f", "-"},
	)

	if err != nil {
		return "`kustomize build` is failed or `kubectl apply` is failed", err
	}

	fmt.Println(string(out))

	return "", nil
}

func Delete(dir string) (string, error) {
	defer removeContents(outputDir)
	fmt.Println("run kustomize build")

	out, err := pipeline.Output(
		[]string{"kustomize", "build", dir},
		[]string{"kubectl", "delete", "-f", "-"},
	)

	if err != nil {
		return "`kustomize build` is failed or `kubectl delete` is failed", err
	}

	fmt.Println(string(out))

	return "", nil
}

func removeContents(dir string) (string, error) {
	d, err := os.Open(dir)
	if err != nil {
		return "Can't Open " + dir, err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return "Can't Read dir names", err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return "Can't Remove All in " + dir, err
		}
	}
	return "", nil
}
