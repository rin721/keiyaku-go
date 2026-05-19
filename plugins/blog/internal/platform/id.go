package platform

import (
	"context"
	"fmt"

	"github.com/bwmarrin/snowflake"
)

type SnowflakeGenerator struct {
	node *snowflake.Node
}

func NewSnowflakeGenerator(nodeID int64) (*SnowflakeGenerator, error) {
	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("create snowflake node: %w", err)
	}
	return &SnowflakeGenerator{node: node}, nil
}

func (g *SnowflakeGenerator) NewID(context.Context) (int64, error) {
	if g == nil || g.node == nil {
		return 0, fmt.Errorf("snowflake generator is not ready")
	}
	return g.node.Generate().Int64(), nil
}
