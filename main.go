package main

import (
	"code.google.com/p/goauth2/oauth"
	"fmt"
	"github.com/google/go-github/github"
	"net/http"
	"os"
	"regexp"
)

var (
	githubClient *github.Client
)

func getRequestedFileContents(r *http.Request) (contents *github.RepositoryContent, err error) {
	var ref, user, repo, filePath string

	url := *r.URL
	path := url.Path

	re := regexp.MustCompile("^/([^/]+)/([^/]+)/(?:blob|raw)/([^/]+)/(.+)$")
	m := re.FindStringSubmatch(path)
	if m == nil {
		return nil, fmt.Errorf("Only URLs matching this pattern are accepted: /{user}/{repo}/(blob|raw)/{ref}/{path}")
	}
	user, repo, ref, filePath = m[1], m[2], m[3], m[4]

	opts := &github.RepositoryContentGetOptions{Ref: ref}
	contents, _, _, err = githubClient.Repositories.GetContents(user, repo, filePath, opts)
	return contents, err
}

func validateRequestVerb(r *http.Request, method string) bool {
	if r.Method != method {
		return false
	}
	return true
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%+v\n", r)
	if !validateRequestVerb(r, "GET") {
		http.Error(w, "GET only", 405)
		return
	}

	fileContents, err := getRequestedFileContents(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	data, err := fileContents.Decode()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, "%s", data)
}

func main() {
	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		fmt.Println("ERROR: Supply the GITHUB_AUTH_TOKEN variable.")
		os.Exit(1)
	}

	clientTransport := &oauth.Transport{Token: &oauth.Token{AccessToken: token}}
	githubClient = github.NewClient(clientTransport.Client())

	http.HandleFunc("/", handler)

	http.ListenAndServe("127.0.0.1:8080", nil)
}
