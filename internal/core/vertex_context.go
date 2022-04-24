package core

import (
	"github.com/MisakiOfScut/go-dage/internal/script"
	"github.com/MisakiOfScut/go-dage/internal/utils/eval"
	"github.com/MisakiOfScut/go-dage/internal/utils/log"
	"go.uber.org/atomic"
	"time"
)

type vertexContext struct {
	id                       string
	operator                 Operator
	result                   int
	remainingDepsNum         atomic.Uint32
	eval                     eval.EvaluableExpression
	nextVertexCtx            []*vertexContext
	depsVertexResult         map[string]int // expected result
	depsVertexesActualResult []int          // store actual result
	depsIdx                  map[string]int

	graphContext *graphContext
}

func newVertexContext(graphContext *graphContext) *vertexContext {
	return &vertexContext{graphContext: graphContext}
}

func (v *vertexContext) isReady() bool {
	return v.remainingDepsNum.Load() == 0
}

// set deps vertex's result and return remaining deps num
// if return val equal to zero, means this vertex is ready to execute
func (v *vertexContext) setDependencyRes(id string, res int) uint32 {
	if _, ok := v.depsIdx[id]; !ok {
		log.Panicf("vertex id:%s not exist in depsIdx map", id)
	}
	idx := v.depsIdx[id]
	latestRes := v.depsVertexesActualResult[idx]
	v.depsVertexesActualResult[idx] = res

	if latestRes == script.VInit {
		return v.remainingDepsNum.Sub(1)
	} else {
		return v.remainingDepsNum.Load()
	}
}

func (v *vertexContext) build(vertex *script.Vertex) {
	if v.operator = v.graphContext.getOprMgr().GetOperator(vertex.Operator); v.operator == nil {
		log.Panicf("vertex id:%s, can't find its operator:%s in operator manager", vertex.ID, vertex.Operator)
	}

	for id, _ := range vertex.NextVertex {
		next := v.graphContext.getVertexCtx(id)
		if next == nil {
			log.Panicf("vertex id:%s, get vertex context from graphContext failed, vertex:%+v, graph:%+v, "+
				"graphContext:%+v", id,
				vertex, v.graphContext.name, v.graphContext)
		}
		v.nextVertexCtx = append(v.nextVertexCtx, next)
	}

	v.depsVertexResult = vertex.DepsVertexResult
	v.depsVertexesActualResult = make([]int, len(vertex.DepsVertexResult))
	v.depsIdx = make(map[string]int)
	idx := 0
	for id := range vertex.DepsVertexResult {
		v.depsIdx[id] = idx
		v.depsVertexesActualResult[idx] = script.VInit
		idx++
	}

	v.id = vertex.ID
	v.eval = vertex.Eval
	v.result = script.VInit
	v.remainingDepsNum.Store(uint32(len(v.depsVertexesActualResult)))
}

func (v *vertexContext) execute() {
	defer v.onFinish()

	// graph execute timeout
	if v.graphContext.getEndTime() != 0 && v.graphContext.getEndTime() <= time.Now().UnixMicro() {
		v.result = script.VAll
		return
	}
	for depVertexId, idx := range v.depsIdx {
		result := v.depsVertexesActualResult[idx]
		expected := v.depsVertexResult[depVertexId]
		if expected == script.VAll {
			continue
		} else if expected != result {
			v.result = script.VFail
			return
		}
	}

	if v.eval != nil {
		v.executeCondProcessor()
	} else {
		v.executeUserProcessor()
	}
}

func (v *vertexContext) executeCondProcessor() {
	v.result = script.VFail
	result, err := v.graphContext.context.DoEval(v.eval)
	if err != nil {
		log.Errorf("vertex:%s, evaluate cond:%s failed with err:%v", v.id, v.eval.String(), err)
		return
	}
	r, ok := result.(bool)
	if !ok {
		log.Errorf("vertex:%s, cond:%s is not a bool expression (its result type isn't bool)", v.id, v.eval.String())
		return
	}

	if r {
		v.result = script.VOk
	}
}

func (v *vertexContext) executeUserProcessor() {
	if err := v.operator.OnExecute(v.graphContext.context); err != nil {
		v.result = script.VFail
		log.Debugf("vertex:%s, with operator:%s, execution return err:%v", v.id, v.operator.Name, err)
	} else {
		v.result = script.VOk
	}
}

func (v *vertexContext) onFinish() {
	v.graphContext.onVertexDone(v)
}

func (v *vertexContext) reset() {
	v.result = script.VInit
	v.remainingDepsNum.Store(uint32(len(v.depsVertexesActualResult)))
	for k, _ := range v.depsVertexesActualResult {
		v.depsVertexesActualResult[k] = script.VInit
	}
}
