package script

import (
	"github.com/BurntSushi/toml"
	"strings"
	"testing"
)

var testScript1 string = `
	[[graph]]
	name = "test_graph_0"

	[[graph.vertex]]
	op = "opr1"
	start = true

	[[graph.vertex]]
	op = "opr2"
	deps = ["opr1"]
	next = ["opr4"]

	[[graph.vertex]]
	op = "opr3"
	deps = ["opr1"]
	next = ["opr4"]

	[[graph.vertex]]
	op = "opr4"
	next = ["cond1"]

	[[graph.vertex]]
	id = "cond1"
	cond = "opr3 > opr2"
	next_on_ok = ["opr5"]
	next_on_fail = ["opr6"]
	
	[[graph.vertex]]
	op = "opr5"

	[[graph.vertex]]
	op = "opr6"
`

type mockGraphManager struct {
}

func (p *mockGraphManager) IsOprExisted(string2 string) bool {
	return true
}

func TestDecode(t *testing.T) {
	gc := NewGraphCluster(&mockGraphManager{})
	if _, err := toml.Decode(testScript1, gc); err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Logf("%+v\n", *gc)
}

func TestBuild(t *testing.T) {
	gc := NewGraphCluster(&mockGraphManager{})
	if _, err := toml.Decode(testScript1, gc); err != nil {
		t.Log(err)
		t.Fail()
	}
	if err := gc.Build(); err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Logf("%+v\n", *gc)
}

func TestGraphCluster_DumpGraphClusterDot(t *testing.T) {
	gc := NewGraphCluster(&mockGraphManager{})
	if _, err := toml.Decode(testScript1, gc); err != nil {
		t.Log(err)
		t.Fail()
	}
	if err := gc.Build(); err != nil {
		t.Log(err)
		t.Fail()
	}
	sb := strings.Builder{}
	gc.DumpGraphClusterDot(&sb)
	t.Log(sb.String())
}

func TestIsolatedVertex(t *testing.T) {
	var testIsolatedVertex = `
[[graph]]
name = "test_graph_0"

[[graph.vertex]]
op = "op1"
start = true

[[graph.vertex]]
op = "op2"
`
	gc := NewGraphCluster(&mockGraphManager{})
	if _, err := toml.Decode(testIsolatedVertex, gc); err != nil {
		t.Log(err)
		t.Fail()
	}
	if err := gc.Build(); err == nil {
		t.Fail()
	} else {
		t.Log(err)
	}
}

func TestCircleCheck(t *testing.T) {
	var testCircleScript = `
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
	gc := NewGraphCluster(&mockGraphManager{})
	if _, err := toml.Decode(testCircleScript, gc); err != nil {
		t.Log(err)
		t.Fail()
	}
	if err := gc.Build(); err != nil {
		t.Log(err)
	} else {
		t.Fail()
	}
}

func TestCondExprParse(t *testing.T) {
	validCond := `
	[[graph]]
	name = "test_cond_parse"

	[[graph.vertex]]
	id = "1"
	cond = "(a * b)/c + d >= 100.00758"
	start = true
`
	gc := NewGraphCluster(&mockGraphManager{})
	if _, err := toml.Decode(validCond, gc); err != nil {
		t.Log(err)
		t.Fail()
	}
	if err := gc.Build(); err != nil {
		t.Log(err)
		t.Fail()
	}
}
