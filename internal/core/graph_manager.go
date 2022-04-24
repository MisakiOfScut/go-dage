package core

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/MisakiOfScut/go-dage/internal/script"
	"github.com/MisakiOfScut/go-dage/internal/utils/executor"
	"github.com/MisakiOfScut/go-dage/internal/utils/log"
	"strings"
	"sync"
)

type dagExecutableGraph struct {
	name          string
	graphClusters *script.GraphCluster
	// graphClusterContextPool executor.ConcurrentQueue
	graphClusterContextPool *sync.Pool
}

func (g *dagExecutableGraph) execute(context *DAGContext, graphName string, timeoutMillisecond int64,
	usersDoneClosure func()) error {
	gc, ok := g.graphClusterContextPool.Get().(*graphClusterContext)
	if !ok {
		log.Panicf("assert from graphClusterContextPool.Get failed")
	}
	return gc.execute(context, graphName, timeoutMillisecond, func() {
		gc.reset()
		g.graphClusterContextPool.Put(gc)
		if usersDoneClosure != nil {
			usersDoneClosure()
		}
	})
}

type GraphManager struct {
	dagGraphs map[string]*dagExecutableGraph
	lock      sync.RWMutex
	executor  executor.Executor
	oprMgr    OperatorManager
}

func NewGraphManager(executor executor.Executor, oprMgr OperatorManager) *GraphManager {
	return &GraphManager{
		dagGraphs: make(map[string]*dagExecutableGraph),
		lock:      sync.RWMutex{},
		executor:  executor,
		oprMgr:    oprMgr,
	}
}

func (m *GraphManager) setDagGraph(dagGraph *dagExecutableGraph) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.dagGraphs[dagGraph.name] = dagGraph
}

func (m *GraphManager) getDagGraph(name string) *dagExecutableGraph {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if v, ok := m.dagGraphs[name]; ok {
		return v
	}
	return nil
}

func (m *GraphManager) IsOprExisted(oprName string) bool {
	return m.oprMgr.GetOperator(oprName) != nil
}

func (m *GraphManager) Execute(userData interface{}, graphClusterName string, graphName string,
	timeoutMillisecond int64, usersDoneClosure func()) error {
	g := m.getDagGraph(graphClusterName)
	if g == nil {
		return fmt.Errorf("graphCluster:%s is not existed", graphClusterName)
	}
	return g.execute(&DAGContext{dagParams: newDagParams(), UserData: userData}, graphName, timeoutMillisecond,
		usersDoneClosure)
}

func (m *GraphManager) Build(dagName string, tomlScript *string) error {
	graphCluster := script.NewGraphCluster(m)
	if _, err := toml.Decode(*tomlScript, graphCluster); err != nil {
		log.Errorf("decode dag:%s failed, %v", dagName, err)
		return err
	}

	if err := graphCluster.Build(); err != nil {
		log.Errorf("build dag:%s failed, %v", dagName, err)
		return err
	}

	dagGraph := &dagExecutableGraph{
		name: dagName, graphClusters: graphCluster, graphClusterContextPool: &sync.Pool{
			New: func() interface{} {
				graphClusterCtx := newGraphClusterContext(m.executor, m.oprMgr)
				graphClusterCtx.build(graphCluster)
				return graphClusterCtx
			},
		},
	}
	m.setDagGraph(dagGraph)

	return nil
}

func (m *GraphManager) Stop() {
	m.executor.Stop()
}

func (m *GraphManager) DumpDAGDot(graphClusterName string) string {
	g := m.getDagGraph(graphClusterName)
	if g == nil {
		return fmt.Sprintf("graphCluster:%s is not existed", graphClusterName)
	}
	sb := strings.Builder{}
	g.graphClusters.DumpGraphClusterDot(&sb)
	return sb.String()
}

func (m *GraphManager) ReplaceExecutor(executor2 executor.Executor) {
	m.executor.Stop()
	m.executor = executor2
}
