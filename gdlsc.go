package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// URL represents path a github repos license
var URL = "https://api.github.com/repos/%s/license?access_token=" + os.Getenv("ACCESS_TOKEN")

//types for message to be printed
type message struct {
	Dependency string
	License    string
	Link       string
}

// options for response from github
type response struct {
	Html_URL string
	License  license
}

type license struct {
	Name string
}

// uses github api to fetch license based on link generated
func getLicense(dependency string, ch chan message) {
	url := formatLink(dependency)
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
	if err := json.Unmarshal([]byte(body), &cleanBody); err != nil {
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

// get just the owner and repo and generate github license link
func formatLink(link string) string {
	strippedURL := strings.Split(link, "/")
	authorAndRepo := strings.Join(strippedURL[1:3], "/")
	licenseURL := fmt.Sprintf(URL, authorAndRepo)
	return licenseURL
}

// removes duplicate links
func removeDuplicates(links []string) []string {
	linksMap := make(map[string]int)
	for _, link := range links {
		linksMap[link] = 1
	}
	linksSlice := make([]string, 0, len(linksMap))
	for link := range linksMap {
		linksSlice = append(linksSlice, link)
	}
	return linksSlice
}

func main() {
	out, err := exec.Command("sh", "-c", `go list -f '{{ join .Imports "\n"}}' ./... | grep github`).Output()
	if err != nil {
		log.Fatal(err)
	}
	depSlice := strings.Split(string(out), "\n")
	rawDeps := depSlice[:len(depSlice)-1] //remove trailing empty string
	dependencies := removeDuplicates(rawDeps)
	ch := make(chan message, len(dependencies)-1)
	for _, rawDep := range rawDeps {
		go getLicense(rawDep, ch)
	}

	for i := 0; i < len(rawDeps); i++ {
		result := <-ch
		fmt.Printf("%v ===>: %v (%v)\n", result.Dependency, result.License, result.Link)
	}
}
