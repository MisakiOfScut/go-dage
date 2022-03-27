package script

import "fmt"

type GraphCluster struct {
	Graph []Graph `toml:"graph"`

	graphMap map[string]*Graph
}

func (gc *GraphCluster) build() error {
	for i := 0; i < len(gc.Graph); i++ {
		g := &gc.Graph[i]
		if err := g.build(); err != nil {
			return err
		}
		if gc.graphMap[g.Name] != nil {
			return fmt.Errorf("[graphCluster] graph %s is duplicated", g.Name)
		}
		gc.graphMap[g.Name] = g
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
		err := v.build()
		if nil != err {
			return err
		}
	}

	// check graph legality

	return nil
}
