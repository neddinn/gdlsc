package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

// URL represents path a github repos license
var URL = "https://api.github.com/repos/%s/license"

type message struct {
	Dependency string
	License    string
	Link       string
}
type response struct {
	Html_URL string
	License  license
}
type license struct {
	Name string
}

func getLicense(url string, dependency string, ch chan message) {
	cleanURL := formatLink(url)
	url = fmt.Sprintf(URL, cleanURL)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	cleanBody := &response{}

	if err := json.Unmarshal([]byte(body), &cleanBody); err!= nil {
		panic(err)
	}

	link := cleanBody.Html_URL
	license := cleanBody.License.Name
	if link == "" {
		link = "Not Found"
	}

	ch <- message{
		Dependency: dependency,
		License:    license,
		Link:       link,
	}
}

// formatLink returns just the repo name and owner in the format ":owner/:repo"
func formatLink(link string) string {
	strippedURL := strings.Split(link, "/")
	cleanURL := strings.Join(strippedURL[1:3], "/")
	return cleanURL
}

func main() {
	out, err := exec.Command("sh", "-c", `go list -f '{{ join .Imports "\n"}}' ./... | grep github`).Output()
	if err != nil {
		log.Fatal(err)
	}
	deps := strings.Split(string(out), "\n")
	rawDeps := deps[:len(deps)-1]
	ch := make(chan message, len(rawDeps)-1)
	for _, rawDep := range rawDeps {
		go getLicense(rawDep, rawDep, ch)
	}
	for i := 0; i < len(rawDeps); i++ {
		result := <-ch
		fmt.Printf("%v ===>: %v (%v)\n", result.Dependency, result.License, result.Link)
	}
}
