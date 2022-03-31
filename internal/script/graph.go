package script

import "fmt"

type GraphCluster struct {
	Graph []Graph `toml:"graph"`

	graphMap map[string]*Graph
}

func (gc *GraphCluster) build() error {
	if gc.graphMap == nil {
		gc.graphMap = make(map[string]*Graph)
	}
	for i := 0; i < len(gc.Graph); i++ {
		g := &gc.Graph[i]
		if gc.graphMap[g.Name] != nil {
			return fmt.Errorf("[graphCluster] graph %s is duplicated", g.Name)
		}
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
		if DFS(vertex, visited) == true {
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

	for _, next := range v.nextVertex {
		if DFS(next, visited) == true {
			return true
		}
	}
	return false
}

func (g *Graph) getVertexByID(id string) *Vertex {
	return g.vertexMap[id]
}
