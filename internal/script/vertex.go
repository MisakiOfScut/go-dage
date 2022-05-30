package script

import (
	"fmt"
	"github.com/MisakiOfScut/go-dage/internal/utils/eval"
	"strings"
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

type Data struct {
	Name string `toml:"name"` // data name
	ID   string `toml:"id"`   // data id (id equals to name by default)
}

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

	Input  []Data `toml:"input"`
	Output []Data `toml:"output"`

	NextVertex       map[string]*Vertex
	DepsVertexResult map[string]int
	Eval             eval.EvaluableExpression
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
	v.NextVertex = make(map[string]*Vertex)
	v.DepsVertexResult = make(map[string]int)

	// cond vertex
	if len(v.Operator) == 0 && len(v.Cond) != 0 {
		if len(v.ID) == 0 {
			return fmt.Errorf("[graph:%s] has a anonymous condition vertex, which must have an ID", v.g.Name)
		}

		// try to get an eval expr from user condition expr
		expression, err := eval.NewEvaluableExpression(v.Cond)
		if err != nil {
			return fmt.Errorf("[graph:%s, vertex id:%s] cond:%s parsed failed with err:%v", v.g.Name,
				v.ID, v.Cond, err)
		}
		v.Eval = expression
		v.Operator = DAGE_EXPR_OPERATOR
		return nil
	}

	// normal vertex
	if len(v.ID) == 0 {
		v.ID = v.Operator
	}
	return v.setUpInputOutput()
}

func (v *Vertex) setUpInputOutput() error {
	if !v.g.GetGraphMgr().IsProduction() {
		return nil
	}

	if !v.g.GetGraphMgr().IsOprExisted(v.Operator) {
		return fmt.Errorf("vertex id:%s, can't find its operator:%s in operator manager", v.ID, v.Operator)
	}

	for _, name := range v.g.GetGraphMgr().GetOperatorInputs(v.Operator) {
		isMatch := false
		for i, _ := range v.Input {
			if v.Input[i].Name == name { // this input has been set by user
				isMatch = true
				break
			}
		}
		if !isMatch {
			v.Input = append(v.Input, Data{Name: name, ID: name})
		}
	}

	for _, name := range v.g.GetGraphMgr().GetOperatorOutputs(v.Operator) {
		isMatch := false
		for i, _ := range v.Output {
			if v.Output[i].Name == name { // this output has been set by user
				isMatch = true
				break
			}
		}
		if !isMatch {
			v.Output = append(v.Output, Data{Name: name, ID: name})
		}
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
	// build vertex's dependencies from data dependencies
	for i, _ := range v.Input {
		preVertex := v.g.getVertexByDataId(v.Input[i].ID)
		if preVertex == nil {
			return fmt.Errorf("[graph:%s, vertex id:%s] can't find vertex input:%s from other vertexes output",
				v.g.Name,
				v.ID, v.Input[i])
		}
		v.depend(preVertex, VOk)
	}

	// build vertex's dependencies from process dependencies
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

// e.x.
// sub_graph2_opr1 [label="opr1" color=black fillcolor=linen style=filled];
// sub_graph2_test_34old [label="user_type==\"34old\"" shape=diamond color=black fillcolor=aquamarine style=filled];
func (v *Vertex) dumpNodeDot(sb *strings.Builder) {
	sb.WriteString(fmt.Sprintf("%s [", v.getDotID()))
	if len(v.Cond) > 0 {
		sb.WriteString(fmt.Sprintf("label=\"%s\" shape=diamond color=black fillcolor=aquamarine style=filled",
			strings.ReplaceAll(v.Cond, "\"", "\\\"")))
	} else {
		sb.WriteString(fmt.Sprintf("label=\"%s\" color=black fillcolor=linen style=filled", v.ID))
	}
	sb.WriteString("];\n")
}

func (v *Vertex) dumpEdgeDot(sb *strings.Builder) {
	if len(v.DepsVertexResult) == 0 {
		// sub_graph2__START__ -> sub_graph2_opr0;
		sb.WriteString(fmt.Sprintf("%s__START__ -> %s;\n", v.g.Name, v.getDotID()))
	}
	if len(v.NextVertex) == 0 {
		// sub_graph2_opr3 -> sub_graph2__STOP__;
		sb.WriteString(fmt.Sprintf("%s -> %s__STOP__;\n", v.getDotID(), v.g.Name))
	}
	for preID, expected := range v.DepsVertexResult {
		sb.WriteString(fmt.Sprintf("%s -> %s ", v.g.GetVertexByID(preID).getDotID(), v.getDotID()))
		switch expected {
		case VOk:
			// sub_graph2_test_34old -> sub_graph2_opr3 [style=dashed label="ok"];
			sb.WriteString("[style=dashed label=\"ok\"];\n")
		case VFail:
			// sub_graph2_test_34old -> sub_graph2_opr4 [style=dashed color=red label="fail"];
			sb.WriteString("[style=dashed color=red label=\"fail\"];\n")
		default:
			// sub_graph2_opr0 -> sub_graph2_test_34old [style=bold label="all"];
			sb.WriteString("[style=bold label=\"all\"];\n")
		}
	}
}

func (v *Vertex) getDotID() string {
	return v.g.Name + "_" + v.ID
}
