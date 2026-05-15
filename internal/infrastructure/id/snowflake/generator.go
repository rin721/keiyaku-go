package snowflake

import (
	"context"
	"fmt"

	"github.com/bwmarrin/snowflake"
)

type Generator struct {
	node *snowflake.Node
}

func New(nodeID int64) (*Generator, error) {
	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("create snowflake node: %w", err)
	}
	return &Generator{node: node}, nil
}

func (g *Generator) NewID(_ context.Context) (int64, error) {
	if g == nil || g.node == nil {
		return 0, fmt.Errorf("snowflake generator is not ready")
	}
	return g.node.Generate().Int64(), nil
}
