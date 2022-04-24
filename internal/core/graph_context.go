package core

import (
	"fmt"
	"github.com/MisakiOfScut/go-dage/internal/script"
	"github.com/MisakiOfScut/go-dage/internal/utils/executor"
	"github.com/MisakiOfScut/go-dage/internal/utils/log"
	"go.uber.org/atomic"
	"time"
)

type graphClusterContext struct {
	endTimeStamp int64 // the timestamp when timeout
	executor     executor.Executor
	oprMgr       OperatorManager
	graphCtxMap  map[string]*graphContext
}

func newGraphClusterContext(executor executor.Executor, oprMgr OperatorManager) *graphClusterContext {
	return &graphClusterContext{
		endTimeStamp: 0,
		executor:     executor,
		oprMgr:       oprMgr,
		graphCtxMap:  make(map[string]*graphContext),
	}
}

func (gc *graphClusterContext) getExecutor() executor.Executor {
	return gc.executor
}

func (gc *graphClusterContext) getOprMgr() OperatorManager {
	return gc.oprMgr
}

func (gc *graphClusterContext) setTimeout(millisecond int64) {
	gc.endTimeStamp = millisecond*1000 + time.Now().UnixMicro()
}

func (gc *graphClusterContext) getEndTime() int64 {
	return gc.endTimeStamp
}

func (gc *graphClusterContext) addGraphCtx(name string, g *graphContext) {
	gc.graphCtxMap[name] = g
}

func (gc *graphClusterContext) build(cluster *script.GraphCluster) {
	for i, _ := range cluster.Graph {
		gc.addGraphCtx(cluster.Graph[i].Name, newGraphContext(gc))
	}
	for name, graphContext := range gc.graphCtxMap {
		graphContext.build(cluster.GetGraphByName(name))
	}
}

func (gc *graphClusterContext) execute(context *DAGContext, graphName string, timeoutMillisecond int64,
	doneClosure func()) error {
	if _, ok := gc.graphCtxMap[graphName]; !ok {
		return fmt.Errorf("graph %s is not existed", graphName)
	}
	if timeoutMillisecond > 0 {
		gc.setTimeout(timeoutMillisecond)
	}

	gc.graphCtxMap[graphName].execute(context, func(){
		gc.endTimeStamp = time.Now().UnixMicro()
		log.Debugf("graph:%s execution ended with Nanoseconds = %v", graphName, gc.getEndTime())
		if doneClosure != nil {
			doneClosure()
		}
	})

	return nil
}

func (gc *graphClusterContext) reset() {
	gc.endTimeStamp = 0
	for _, graphCtx := range gc.graphCtxMap {
		graphCtx.reset()
	}
}

type graphContext struct {
	name              string
	remainingVertexes atomic.Uint32
	vertexCtxMap      map[string]*vertexContext

	// runtime assign
	context     *DAGContext
	doneClosure func()

	graphClusterCtx *graphClusterContext
}

func newGraphContext(ctx *graphClusterContext) *graphContext {
	return &graphContext{graphClusterCtx: ctx, vertexCtxMap: make(map[string]*vertexContext)}
}

func (g *graphContext) getEndTime() int64 {
	return g.graphClusterCtx.getEndTime()
}

func (g *graphContext) getOprMgr() OperatorManager {
	return g.graphClusterCtx.getOprMgr()
}

func (g *graphContext) getVertexCtx(id string) *vertexContext {
	return g.vertexCtxMap[id]
}

func (g *graphContext) addVertexCtx(id string, v *vertexContext) {
	g.vertexCtxMap[id] = v
}

func (g *graphContext) build(graph *script.Graph) {
	for i, _ := range graph.Vertex {
		g.addVertexCtx(graph.Vertex[i].ID, newVertexContext(g))
	}
	for id, vertexContext := range g.vertexCtxMap {
		vertexContext.build(graph.GetVertexByID(id))
	}
	g.remainingVertexes.Store(uint32(len(g.vertexCtxMap)))
	g.name = graph.Name
}

func (g *graphContext) execute(context *DAGContext, doneClosure func()) {
	g.context = context
	g.doneClosure = doneClosure

	var readyVertex []*vertexContext
	for _, vertexCtx := range g.vertexCtxMap {
		if vertexCtx.isReady() {
			readyVertex = append(readyVertex, vertexCtx)
		}
	}
	g.executeReadyVertex(readyVertex)
}

func (g *graphContext) executeReadyVertex(vertexes []*vertexContext) {
	for _, v := range vertexes {
		vCatch := v
		g.graphClusterCtx.getExecutor().Execute(func() {
			vCatch.execute()
		})
	}
}

func (g *graphContext) onVertexDone(v *vertexContext) {
	if g.remainingVertexes.Sub(1) == 0 {
		g.doneClosure()
		return
	}

	var readyVertex []*vertexContext
	for i, _ := range v.nextVertexCtx {
		if v.nextVertexCtx[i].setDependencyRes(v.id, v.result) == 0 {
			readyVertex = append(readyVertex, v.nextVertexCtx[i])
		}
	}
	g.executeReadyVertex(readyVertex)
}

func (g *graphContext) reset() {
	g.context = nil
	g.remainingVertexes.Store(uint32(len(g.vertexCtxMap)))
	for _, vertexContext := range g.vertexCtxMap {
		vertexContext.reset()
	}
}
