package script

import (
	"fmt"
	"strings"
)

type GraphCluster struct {
	Graph []Graph `toml:"graph"`

	isBuild  bool
	graphMgr IGraphManager
	graphMap map[string]*Graph
}

type IGraphManager interface {
	IsOprExisted(oprName string) bool
	GetOperatorInputs(oprName string) []string
	GetOperatorOutputs(oprName string) []string
	IsProduction() bool
}

func NewGraphCluster(gMgr IGraphManager) *GraphCluster {
	return &GraphCluster{graphMgr: gMgr, isBuild: false, graphMap: make(map[string]*Graph)}
}

func (gc *GraphCluster) Build() error {
	for i := 0; i < len(gc.Graph); i++ {
		g := &gc.Graph[i]
		if gc.graphMap[g.Name] != nil {
			return fmt.Errorf("graph %s is duplicated", g.Name)
		}
		g.cluster = gc
		gc.graphMap[g.Name] = g

		if err := g.build(); err != nil {
			return err
		}
		if err := g.verifyAfterBuild(); err != nil {
			return err
		}
	}
	gc.isBuild = true
	return nil
}

func (gc *GraphCluster) IsBuild() bool {
	return gc.isBuild
}

func (gc *GraphCluster) DumpGraphClusterDot(sb *strings.Builder) {
	if gc.isBuild == false {
		sb.WriteString(fmt.Sprintf("graphCluster is not built yet"))
		return
	}

	sb.WriteString("digraph G {\nrankdir=LR;\n")
	for _, graph := range gc.graphMap {
		graph.dumpDot(sb)
	}
	sb.WriteString("}\n")
}

func (gc *GraphCluster) GetGraphByName(name string) *Graph {
	if g, ok := gc.graphMap[name]; ok {
		return g
	}
	return nil
}

func (gc *GraphCluster) GetGraphMgr() IGraphManager {
	return gc.graphMgr
}

type Graph struct {
	Name   string   `toml:"name"`
	Vertex []Vertex `toml:"vertex"`

	cluster       *GraphCluster
	vertexMap     map[string]*Vertex // map vertex id to *Vertex
	OutputDataMap map[string]*Vertex // map output data id to *Vertex
}

func (g *Graph) build() error {
	if g.vertexMap == nil {
		g.vertexMap = make(map[string]*Vertex)
	}
	if g.OutputDataMap == nil {
		g.OutputDataMap = make(map[string]*Vertex)
	}

	for i := 0; i < len(g.Vertex); i++ {
		v := &g.Vertex[i]
		v.g = g
		if err := v.verifyAndSetUp(); err != nil {
			return err
		}
		if g.vertexMap[v.ID] != nil {
			return fmt.Errorf("[graph:%s] vertex %s is duplicated", g.Name, v.ID)
		}
		g.vertexMap[v.ID] = v
	}

	if err := g.buildInputOutput(); err != nil {
		return err
	}

	// build vertexes' dependency
	for _, v := range g.vertexMap {
		if err := v.build(); err != nil {
			return err
		}
	}

	return nil
}

func (g *Graph) buildInputOutput() error {
	for _, v := range g.vertexMap {
		for i, _ := range v.Output {
			if t := g.getVertexByDataId(v.Output[i].ID); t != nil {
				return fmt.Errorf("[graph:%s] vertex:%s and vertex:%s have a duplicated data:%v in output", v.g.Name,
					v.ID,
					t.ID, v.Output[i])
			}
			g.OutputDataMap[v.Output[i].ID] = v
		}
	}
	// for _, v := range g.vertexMap {
	// 	for i, _ := range v.Input {
	// 		if t := g.getVertexByDataId(v.Input[i].ID); t == nil {
	// 			return fmt.Errorf("[graph:%s, vertex id:%s] can't find vertex input:%v from other vertexes output",
	// 				v.g.Name,
	// 				v.ID, v.Input[i])
	// 		}
	// 		g.inputDataMap[v.Input[i].ID] = append(g.inputDataMap[v.Input[i].ID], v)
	// 	}
	// }

	return nil
}

func (g *Graph) getVertexByDataId(dataId string) *Vertex {
	if val, existed := g.OutputDataMap[dataId]; existed {
		return val
	}
	return nil
}

func (g *Graph) dumpDot(sb *strings.Builder) {
	sb.WriteString(fmt.Sprintf("subgraph cluster_%s {\n", g.Name))
	sb.WriteString("style = rounded;\n")
	sb.WriteString(fmt.Sprintf("label = \"%s\";\n", g.Name))
	sb.WriteString(g.Name + "__START__[color=black fillcolor=deepskyblue style=filled shape=Msquare" +
		" label=\"START\"];\n")
	sb.WriteString(g.Name + "__STOP__[color=black fillcolor=deepskyblue style=filled shape=Msquare" +
		" label=\"STOP\"];\n")
	for _, vertex := range g.vertexMap {
		vertex.dumpNodeDot(sb)
	}
	sb.WriteString("\n")
	for _, vertex := range g.vertexMap {
		vertex.dumpEdgeDot(sb)
	}
	sb.WriteString("};\n")
}

// check graph legality
func (g *Graph) verifyAfterBuild() error {
	for _, v := range g.vertexMap {
		if err := v.verifyAfterBuild(); err != nil {
			return err
		}
	}
	// check circle
	if g.checkCircle() == true {
		return fmt.Errorf("[graph:%s] has a circle", g.Name)
	}
	return nil
}

func (g *Graph) checkCircle() bool {
	for _, vertex := range g.vertexMap {
		visited := make(map[string]bool)
		if DFS(vertex, visited) {
			return true
		}
	}
	return false
}

func DFS(v *Vertex, visited map[string]bool) bool {
	if _, exist := visited[v.ID]; exist {
		return true
	}
	visited[v.ID] = true

	for _, next := range v.NextVertex {
		visitedCp := make(map[string]bool)
		for k, v := range visited {
			visitedCp[k] = v
		}
		if DFS(next, visitedCp) == true {
			return true
		}
	}
	return false
}

func (g *Graph) GetVertexByID(id string) *Vertex {
	if v, ok := g.vertexMap[id]; ok {
		return v
	}
	return nil
}

func (g *Graph) GetGraphMgr() IGraphManager {
	return g.cluster.GetGraphMgr()
}
