package main

import (
	"log"
	"os/exec"
	"strings"
)

// import "github.com/mitchellh/gox"
// import "golang.org/x/tools/cmd/stringer"

func main() {
	go_bin, err := exec.LookPath("go")
	if err != nil {
		log.Fatal("Go not found in PATH, install it e.g. via 'brew install go'")
	}

	args := []string{
		"list", "-f", "{{join .Deps \" \"}}",
	}
	deps, err := exec.Command(go_bin, args...).Output()
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, pkg := range strings.Fields(string(deps)) {
		log.Println(pkg)
	}
}

// @LIST=( )
// 	go list ./... \
// 		| xargs git --git-dir=$GOPATH/src/{}/.git --work-tree=$GOPATH/src/{} log -1
