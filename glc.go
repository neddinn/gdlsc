package main

import (
  "log"
  "fmt"
  "os/exec"
  "strings"
  "net/http"
  "io/ioutil"
)

var URL = "https://%s/master/LICENSE"

type response struct{
  Dep string
  License string
  Link string
}

func getLicense(url string, dependency string, ch chan response) {
  resp, err := http.Get(url)
  if err != nil {
      panic(err)
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
      panic(err)
  }
  license := getLicenseText(string(body))
  ch <- response{
    Dep: dependency,
    License: license,
    Link: url,
  }
}

func getLicenseText(data string) string {
  whole := strings.Split(data, "\n")
  return whole[0]
}

func main() {
  out, err := exec.Command("sh", "-c", `go list -f '{{ join .Imports "\n"}}' | grep github`).Output()
  if err != nil {
    log.Fatal(err)
  }
  deps := strings.Split(string(out), "\n")
  ch := make(chan response, len(deps)- 1)
  rawDeps :=  deps[:len(deps)-1]
  for _, rawDep := range rawDeps {
    dep := strings.Replace(rawDep, "github.com", "raw.githubusercontent.com", 1)
    dep = fmt.Sprintf(URL, dep)
    go getLicense(dep, rawDep, ch)
  }
  for i := 0; i < len(rawDeps); i++  {
    result := <- ch
    fmt.Printf("%v ===>: %v\n", result.Dep, result.License)
  }
}
