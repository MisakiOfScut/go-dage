package script

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"testing"
)

var testScript1 string = `
[[graph]]
name = "test_graph_0"

[[graph.vertex]]
op = "opr1"
start = true

[[graph.vertex]]
id = "opr2"
cond = "true"
deps = ["opr1"]
next = ["opr3"]

[[graph.vertex]]
op = "opr3"

[[graph]]
name = "test_graph_1"
`

func TestDecode(t *testing.T) {
	gc := &GraphCluster{}
	if _, err := toml.Decode(testScript1, gc); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", *gc)
}

func TestBuild(t *testing.T) {
	gc := &GraphCluster{}
	if _, err := toml.Decode(testScript1, gc); err != nil {
		t.Fatal(err)
	}
	if err := gc.build(); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", *gc)
}

var testIsolatedVertex string = `
[[graph]]
name = "test_graph_0"

[[graph.vertex]]
op = "op1"
start = true

[[graph.vertex]]
op = "op2"
`

func TestIsolatedVertex(t *testing.T) {
	gc := &GraphCluster{}
	if _, err := toml.Decode(testIsolatedVertex, gc); err != nil {
		t.Fatal(err)
	}
	if err := gc.build(); err == nil {
		t.Fail()
	} else {
		t.Log(err)
	}
}

var testCircleScript string = `
[[graph]]
name = "test_circle"

[[graph.vertex]]
op = "opr0"
start = true

[[graph.vertex]]
op = "opr1"
deps = ["opr0"]

[[graph.vertex]]
op = "opr2"
deps = ["opr1"]
next = ["opr3"]

[[graph.vertex]]
op = "opr3"
next = ["opr4"]

[[graph.vertex]]
op = "opr4"
next = ["opr1"]
`

func TestCircleCheck(t *testing.T) {
	gc := &GraphCluster{}
	if _, err := toml.Decode(testCircleScript, gc); err != nil {
		t.Fatal(err)
	}
	if err := gc.build(); err != nil {
		t.Log(err)
	} else {
		t.Fail()
	}
}
