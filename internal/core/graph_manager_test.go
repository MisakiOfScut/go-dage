package core

import (
	"fmt"
	"github.com/MisakiOfScut/go-dage/internal/utils/executor"
	"github.com/MisakiOfScut/go-dage/internal/utils/log"
	"go.uber.org/zap"
	"testing"
	"time"
)

func init() {
	logger := zap.NewExample()
	defer logger.Sync()
	log.SetLogger(logger.Sugar())
}

func TestCreateGraphMgr(t *testing.T) {
	NewGraphManager(executor.NewDefaultExecutor(32, 8), NewDefaultOperatorManager())
}

var gMgr *GraphManager

const graphClusterName = "test_graphCluster_0"
const graphName = "test_graph_0"

var tomlScript0 = `
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

func TestGraphManager_Build(t *testing.T) {
	TestNewDefaultOperatorManager(t)
	gMgr = NewGraphManager(executor.NewDefaultExecutor(32, 8), tOprMgr)
	if err := gMgr.Build(graphClusterName, &tomlScript0); err != nil {
		fmt.Println(err)
		t.Fail()
	}
}

func TestGraphManager_Execute(t *testing.T) {
	TestGraphManager_Build(t)
	if err := gMgr.Execute(nil, graphClusterName, graphName, 0); err != nil {
		fmt.Println(err)
		t.Fail()
	}
	time.Sleep(1 * time.Millisecond)
}
