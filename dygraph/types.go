package dygraph

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
)

type Node struct {
	OrganisationId uuid.UUID
	Id             uuid.UUID
	Type           string
	Data           string
}

type LogicalRecord struct {
	Node
	TypeTarget []string
}

type Edge struct {
	OrganisationId uuid.UUID
	Id             uuid.UUID
	TargetNodeId   uuid.UUID
	TargetNodeType string
	Tags           []string
	Data           string
}

const separator = "|"
const nodePrefix = "node_"
const edgePrefix = "edge_"

func (node *Node) createNodeDto() *dto {
	return &dto{
		GlobalId:       fmt.Sprintf("%s_%s", node.OrganisationId, node.Id),
		TypeTarget:     nodePrefix + strings.Join([]string{node.Type, node.Id.String()}, separator),
		OrganisationId: node.OrganisationId.String(),
		Id:             node.Id.String(),
		Type:           node.Type,
		Data:           node.Data,
	}
}

func (r *Edge) createEdgeDto() *dto {
	d := &dto{
		GlobalId:       fmt.Sprintf("%s_%s", r.OrganisationId, r.Id),
		TypeTarget:     edgePrefix + r.TargetNodeType + separator + r.TargetNodeId.String(),
		OrganisationId: r.OrganisationId.String(),
		Id:             r.Id.String(),
		Type:           r.TargetNodeType,
		Data:           r.Data,
	}
	//if d.Data == "" {
	//	d.Data = r.TargetNodeId.String()
	//}

	if r.Tags != nil {
		d.TypeTarget += separator + strings.Join(r.Tags, separator)
	}

	return d
}
