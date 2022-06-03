package main

import (
	"flag"
	"fmt"
	"github.com/MisakiOfScut/go-dage"
	"io/ioutil"
	"os/exec"
	"path"
)

type mockGraphManager struct {
}

func (p *mockGraphManager) IsOprExisted(string2 string) bool {
	return true
}

// go run gen.go -toml=xxx
func main() {
	scriptPath := flag.String("toml", "example.toml", "Specify input toml script")
	flag.Parse()

	if len(*scriptPath) == 0 {
		flag.Usage()
		return
	}

	bytes, err := ioutil.ReadFile(*scriptPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	tomlScript := string(bytes)

	dot, err := dage.TestBuildDAG(&tomlScript)
	if err != nil {
		fmt.Println(err)
		return
	}

	dotFile := path.Base(*scriptPath) + ".dot"
	err = ioutil.WriteFile(dotFile, []byte(dot), 0755)
	if nil != err {
		fmt.Println(err)
		return
	}
	pngFile := path.Base(*scriptPath) + ".png"
	_, err = exec.Command("dot", "-Tpng", dotFile, "-o", pngFile).Output()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Write png into %s\n", pngFile)
}
