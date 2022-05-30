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

type graphExecutor struct {
	name          string
	graphClusters *script.GraphCluster
	// graphClusterContextPool executor.ConcurrentQueue
	graphClusterContextPool *sync.Pool
}

func (g *graphExecutor) execute(context *DAGContext, graphName string, timeoutMillisecond int64,
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
	graphExecutors map[string]*graphExecutor
	lock           sync.RWMutex
	taskExecutor   executor.Executor
	oprMgr         OperatorManager
}

func NewGraphManager(executor executor.Executor, oprMgr OperatorManager) *GraphManager {
	return &GraphManager{
		graphExecutors: make(map[string]*graphExecutor),
		lock:           sync.RWMutex{},
		taskExecutor:   executor,
		oprMgr:         oprMgr,
	}
}

func (m *GraphManager) setGraphExecutor(ge *graphExecutor) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.graphExecutors[ge.name] = ge
}

func (m *GraphManager) getGraphExecutor(name string) *graphExecutor {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if v, ok := m.graphExecutors[name]; ok {
		return v
	}
	return nil
}

func (m *GraphManager) IsOprExisted(oprName string) bool {
	return m.oprMgr.GetOperator(oprName) != nil
}

func (m *GraphManager) GetOperatorInputs(oprName string) []string {
	return m.oprMgr.GetOperator(oprName).GetInputsID()
}

func (m *GraphManager) GetOperatorOutputs(oprName string) []string {
	return m.oprMgr.GetOperator(oprName).GetOutputsID()
}

func (m *GraphManager) IsProduction() bool {
	return true
}

func (m *GraphManager) Execute(userData interface{}, graphClusterName string, graphName string,
	timeoutMillisecond int64, usersDoneClosure func()) error {
	g := m.getGraphExecutor(graphClusterName)
	if g == nil {
		return fmt.Errorf("graphCluster:%s is not existed", graphClusterName)
	}
	return g.execute(&DAGContext{dagParams: newDagParams(), UserData: userData}, graphName, timeoutMillisecond,
		usersDoneClosure)
}

func (m *GraphManager) Build(clusterName string, tomlScript *string) error {
	graphCluster := script.NewGraphCluster(m)
	if _, err := toml.Decode(*tomlScript, graphCluster); err != nil {
		log.Errorf("decode dag:%s failed, %v", clusterName, err)
		return err
	}

	if err := graphCluster.Build(); err != nil {
		log.Errorf("build dag:%s failed, %v", clusterName, err)
		return err
	}

	ge := &graphExecutor{
		name: clusterName, graphClusters: graphCluster, graphClusterContextPool: &sync.Pool{
			New: func() interface{} {
				graphClusterCtx := newGraphClusterContext(m.taskExecutor, m.oprMgr)
				graphClusterCtx.build(graphCluster)
				return graphClusterCtx
			},
		},
	}
	m.setGraphExecutor(ge)

	return nil
}

func (m *GraphManager) Stop() {
	m.taskExecutor.Stop()
}

func (m *GraphManager) DumpDAGDot(graphClusterName string) string {
	g := m.getGraphExecutor(graphClusterName)
	if g == nil {
		return fmt.Sprintf("graphCluster:%s is not existed", graphClusterName)
	}
	sb := strings.Builder{}
	g.graphClusters.DumpGraphClusterDot(&sb)
	return sb.String()
}

func (m *GraphManager) ReplaceTaskExecutor(executor2 executor.Executor) {
	m.taskExecutor.Stop()
	m.taskExecutor = executor2
}
