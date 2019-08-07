package main

import (
	"bytes"
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
	applyPath   = "/apply"
	destroyPath = "/destroy"
	outputDir   = "./tmp"
)

type Config struct {
	BranchName    string
	RepoUrl       string
	GitUser       string
	PersonalToken string
	TerraformPath string
	Port          string
	RootDir       string
}

type Message struct {
	Id               string `json:"id"`
	RepoUrl          string `json:"repoUrl"`
	BranchName       string `json:"branchName"`
	KustomizationUrl string `json:"kustomizationUrl"`
}

var envConfig Config

func main() {
	err := setConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	envConfig.Port = ":80"

	fmt.Println("start terraforming at" + envConfig.Port)

	http.HandleFunc(applyPath, applyRequest)
	http.HandleFunc(destroyPath, destroyRequest)

	http.ListenAndServe(envConfig.Port, nil)
}

func setConfig() error {
	errString := ""
	envConfig.GitUser = os.Getenv("GIT_USER")
	envConfig.PersonalToken = os.Getenv("PERSONAL_TOKEN")
	envConfig.TerraformPath = os.Getenv("TERRAFORM_PATH")
	envConfig.RootDir, _ = os.Getwd()
	if envConfig.GitUser == "" {
		errString = errString + "GIT_USER"
	}
	if envConfig.PersonalToken == "" {
		errString = errString + " PERSONAL_TOKEN"
	}
	if envConfig.TerraformPath == "" {
		errString = errString + " TERRAFORM_PATH"
	}
	if errString != "" {
		return errors.New(errString + " are not set")
	}
	return nil
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
	// '{"id":1234,"repoUrl":"https://github.com/username/reponame","branchName":"feature/hogehoge",
	//   "kustomizationUrl":"http://hogehoge"}'
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if isError("", err, w, 500) {
		fmt.Println(r.Body)
		return
	}

	var msg Message
	err = json.Unmarshal(b, &msg)
	if isError("", err, w, 500) {
		return
	}

	errMessage, err := checkout(msg.RepoUrl, msg.BranchName, outputDir, envConfig.GitUser, envConfig.PersonalToken)
	if isError(errMessage, err, w, 500) {
		return
	}

	errMessage, err = generateVariables(outputDir+envConfig.TerraformPath, msg.Id)
	if isError(errMessage, err, w, 500) {
		return
	}

	errMessage, err = plan(outputDir + envConfig.TerraformPath)
	if isError(errMessage, err, w, 500) {
		return
	}

	errMessage, err = apply(outputDir + envConfig.TerraformPath)
	fmt.Println(errMessage)

	os.Chdir(envConfig.RootDir)
	removeContents(outputDir)

	newAppGenerate(msg.Id, msg.RepoUrl, msg.BranchName, msg.KustomizationUrl)

	output, err := json.Marshal(msg)
	if isError("", err, w, 500) {
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
}

func destroyRequest(w http.ResponseWriter, r *http.Request) {
	// '{"id":1234,"repoUrl":"https://github.com/username/reponame","branchName":"feature/hogehoge"}'
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if isError("", err, w, 500) {
		return
	}

	var msg Message
	err = json.Unmarshal(b, &msg)
	if isError("", err, w, 500) {
		return
	}

	errMessage, err := checkout(msg.RepoUrl, msg.BranchName, outputDir, envConfig.GitUser, envConfig.PersonalToken)
	if isError(errMessage, err, w, 500) {
		return
	}

	errMessage, err = generateVariables(outputDir+envConfig.TerraformPath, msg.Id)
	if isError(errMessage, err, w, 500) {
		return
	}

	errMessage, err = destroy(outputDir + envConfig.TerraformPath)
	if isError(errMessage, err, w, 500) {
		return
	}

	os.Chdir(envConfig.RootDir)
	removeContents(outputDir)

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
			Username: user, // yes, this can be anything except an empty string
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
	fmt.Println("change " + dir + "/variables.tf")
	_, err := pipeline.Output(
		[]string{"sed", "-i", "-e", "s/dev/dev-" + id + "/g", dir + "/variables.tf"},
	)
	if err != nil {
		return "Change environment name is failed in variables.tf", err
	}

	fmt.Println("change " + dir + "/init.tf")
	_, err = pipeline.Output(
		[]string{"sed", "-i", "-e", "s/dev/dev-" + id + "/g", dir + "/init.tf"},
	)
	if err != nil {
		return "Change environment name is failed in init.tf", err
	}

	fmt.Println("change " + dir + " terraform resources")
	_, err = pipeline.Output(
		[]string{"find", ".", "-type", "f", "-name", "*.tf", "-print0"},
		[]string{"xargs", "-0", "sed", "-i", "-e", "s/dev_/dev_" + id + "_/"},
	)
	if err != nil {
		return "Change environment name is failed", err
	}

	return "", nil
}

func plan(dir string) (string, error) {
	fmt.Println("run terraform init")
	os.Chdir(dir)

	out, err := pipeline.Output(
		[]string{"terraform", "init"},
	)
	if err != nil {
		return "`terraform init` is failed", err
	}
	fmt.Println(string(out))

	fmt.Println("run terraform plan")

	out, err = pipeline.Output(
		[]string{"terraform", "plan"},
	)
	if err != nil {
		return "`terraform plan` is failed", err
	}
	fmt.Println(string(out))

	return "", nil
}

func apply(dir string) (string, error) {
	fmt.Println("run terraform apply")
	os.Chdir(dir)

	out, _ := pipeline.Output(
		[]string{"terraform", "apply", "-auto-approve"},
	)
	fmt.Println(string(out))

	fmt.Println("rerun terraform apply")
	out, err := pipeline.Output(
		[]string{"terraform", "apply", "-auto-approve"},
	)
	if err != nil {
		fmt.Println(err)
		return "`terraform apply` is failed", err
	}
	fmt.Println(string(out))

	return "", nil
}

func destroy(dir string) (string, error) {
	fmt.Println("run terraform init")
	os.Chdir(dir)

	out, err := pipeline.Output(
		[]string{"terraform", "init"},
	)
	if err != nil {
		return "terraform init is failed", err
	}
	fmt.Println(string(out))

	fmt.Println("run terraform destroy")
	out, _ = pipeline.Output(
		[]string{"terraform", "destroy", "-auto-approve"},
	)
	fmt.Println(string(out))

	fmt.Println("rerun terraform destroy")
	out, err = pipeline.Output(
		[]string{"terraform", "destroy", "-auto-approve"},
	)
	if err != nil {
		return "`terraform destroy` is failed", err
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
	fmt.Println("remove complete")

	return "", nil
}

func newAppGenerate(id, repoUrl, branchName, url string) (string, error) {
	jsonStr := `{"id":"` + id + `","repoURL":"` + repoUrl + `","branchName":"` + branchName + `"}`

	req, err := http.NewRequest(
		"POST",
		"http://"+url,
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		return "Failed to create http request", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "Request is failed", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf("statusCode: %d, Body: %+v\n", resp.StatusCode, resp.Body)
	}
	return "", nil
}
