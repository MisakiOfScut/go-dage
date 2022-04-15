package script

import (
	"fmt"
)

const (
	VInit = 0
	VOk   = 1
	VFail = 2
	VAll  = 3
)

const (
	DAGE_EXPR_OPERATOR string = "__DAGE_EXPR_OPERATOR__"
)

type Vertex struct {
	ID       string `toml:"id"`
	Operator string `toml:"op"`
	Start    bool   `toml:"start"`
	// Expected string `toml:"expected"`
	Cond string `toml:"cond"`

	Next       []string `toml:"next"`
	NextOnOk   []string `toml:"next_on_ok"`
	NextOnFail []string `toml:"next_on_fail"`
	Deps       []string `toml:"deps"`
	DepsOnOk   []string `toml:"deps_on_ok"`
	DepsOnFail []string `toml:"deps_on_fail"`

	NextVertex       map[string]*Vertex
	DepsVertexResult map[string]int
	g                *Graph
}

func (v *Vertex) verifyAndSetUp() error {
	if len(v.Operator) != 0 && len(v.Cond) != 0 {
		return fmt.Errorf("[graph:%s] vertex id:%s operator:%s cond:%s, "+
			"a vertex can't have operator and cond at the same time", v.g.Name, v.ID, v.Operator, v.Cond)
	}
	if len(v.Operator) == 0 && len(v.Cond) == 0 {
		return fmt.Errorf("[graph:%s] has an anonymous vertex, there are one or more "+
			"normal vertexes haven't operator (or one or more condition vertexes haven't ID)", v.g.Name)
	}

	if v.NextVertex == nil {
		v.NextVertex = make(map[string]*Vertex)
	}
	if v.DepsVertexResult == nil {
		v.DepsVertexResult = make(map[string]int)
	}

	// cond vertex
	if len(v.Operator) == 0 && len(v.Cond) != 0 {
		if len(v.ID) == 0 {
			return fmt.Errorf("[graph:%s] has a anonymous condition vertex, which must have an ID", v.g.Name)
		}
		v.Operator = DAGE_EXPR_OPERATOR
		return nil
	}

	// normal vertex
	if len(v.ID) == 0 {
		v.ID = v.Operator
	}
	if v.g.GetGraphMgr().IsOprExisted(v.Operator) == false {
		return fmt.Errorf("[graph:%s, vertex id:%s] can't find its operator:%s in operator manager", v.g.Name,
			v.ID, v.Operator)
	}

	return nil
}

func (v *Vertex) depend(pre *Vertex, expectedResult int) {
	// if v.DepsVertexResult == nil {
	// 	v.DepsVertexResult = make(map[string]int)
	// }
	// if pre.NextVertex == nil {
	// 	pre.NextVertex = make(map[string]*Vertex)
	// }

	// pre --> v
	v.DepsVertexResult[pre.ID] = expectedResult
	pre.NextVertex[v.ID] = v
}

func (v *Vertex) build() error {
	for _, nextVertexID := range v.Next {
		if nextVertex := v.g.GetVertexByID(nextVertexID); nextVertex != nil {
			nextVertex.depend(v, VAll)
		} else {
			return fmt.Errorf("[graph:%s, vertex id:%s] in vertex's next array, id:%s is not existed", v.g.Name,
				v.ID, nextVertexID)
		}
	}
	for _, nextVertexID := range v.NextOnOk {
		if nextVertex := v.g.GetVertexByID(nextVertexID); nextVertex != nil {
			nextVertex.depend(v, VOk)
		} else {
			return fmt.Errorf("[graph:%s, vertex id:%s] in vertex's next_on_ok array, id:%s is not existed", v.g.Name,
				v.ID, nextVertexID)
		}
	}
	for _, nextVertexID := range v.NextOnFail {
		if nextVertex := v.g.GetVertexByID(nextVertexID); nextVertex != nil {
			nextVertex.depend(v, VFail)
		} else {
			return fmt.Errorf("[graph:%s, vertex id:%s] in vertex's next_on_fail array, id:%s is not existed", v.g.Name,
				v.ID, nextVertexID)
		}
	}
	for _, preVertexID := range v.Deps {
		if preVertex := v.g.GetVertexByID(preVertexID); preVertex != nil {
			v.depend(preVertex, VAll)
		} else {
			return fmt.Errorf("[graph:%s, vertex id:%s] in vertex's deps array, id:%s is not existed", v.g.Name,
				v.ID, preVertexID)
		}
	}
	for _, preVertexID := range v.DepsOnOk {
		if preVertex := v.g.GetVertexByID(preVertexID); preVertex != nil {
			v.depend(preVertex, VOk)
		} else {
			return fmt.Errorf("[graph:%s, vertex id:%s] in vertex's deps_on_ok array, id:%s is not existed", v.g.Name,
				v.ID, preVertexID)
		}
	}
	for _, preVertexID := range v.DepsOnFail {
		if preVertex := v.g.GetVertexByID(preVertexID); preVertex != nil {
			v.depend(preVertex, VFail)
		} else {
			return fmt.Errorf("[graph:%s, vertex id:%s] in vertex's deps_on_fail array, id:%s is not existed", v.g.Name,
				v.ID, preVertexID)
		}
	}

	return nil
}

func (v *Vertex) verifyAfterBuild() error {
	// 1. start vertex shouldn't have deps
	// 2. non-start vertex shouldn't be isolated
	if v.Start {
		if len(v.DepsVertexResult) > 0 {
			return fmt.Errorf("[graph:%s, vertex id:%s] a start vertex shouldn't have deps vertex, "+
				"delete the start flag or deps array",
				v.g.Name,
				v.ID)
		}
	} else if len(v.DepsVertexResult) == 0 && len(v.NextVertex) == 0 {
		return fmt.Errorf("[graph:%s, vertex id:%s] non-start vertex shouldn't be isolated",
			v.g.Name,
			v.ID)
	}
	return nil
}
