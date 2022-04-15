package script

import (
	"fmt"
)

type GraphCluster struct {
	Graph []Graph `toml:"graph"`

	graphMgr IGraphManager
	graphMap map[string]*Graph
}

type IGraphManager interface {
	IsOprExisted(oprName string) bool
}

func NewGraphCluster(gMgr IGraphManager) *GraphCluster {
	return &GraphCluster{graphMgr: gMgr, graphMap: make(map[string]*Graph)}
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
	return nil
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

	cluster   *GraphCluster
	vertexMap map[string]*Vertex
}

func (g *Graph) build() error {
	if g.vertexMap == nil {
		g.vertexMap = make(map[string]*Vertex)
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

	// build vertexes' dependency
	for _, v := range g.vertexMap {
		if err := v.build(); err != nil {
			return err
		}
	}

	return nil
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
		for k,v := range visited{
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
