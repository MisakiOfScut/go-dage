package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/MisakiOfScut/go-dage/internal/script"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
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

	gc := script.NewGraphCluster(&mockGraphManager{})
	if _, err := toml.DecodeFile(*scriptPath, gc); err != nil {
		fmt.Println(err)
		return
	}
	if err := gc.Build(); err != nil {
		fmt.Println(err)
		return
	}
	sb := strings.Builder{}
	gc.DumpGraphClusterDot(&sb)

	dot := sb.String()
	dotFile := path.Base(*scriptPath) + ".dot"
	err := ioutil.WriteFile(dotFile, []byte(dot), 0755)
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
