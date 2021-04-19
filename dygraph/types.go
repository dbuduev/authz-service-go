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

type Edge struct {
	OrganisationId uuid.UUID
	Id             uuid.UUID
	TargetNodeId   uuid.UUID
	TargetNodeType string
	Tags           []string
	Data           string
}

type dto struct {
	GlobalId       string `dynamodbav:"globalId"`
	TypeTarget     string `dynamodbav:"typeTarget"`
	OrganisationId string `dynamodbav:"organisationId"`
	Id             string `dynamodbav:"id"`
	Type           string `dynamodbav:"type"`
	Data           string `dynamodbav:"data"`
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

func (d *dto) createNode() Node {
	return Node{
		OrganisationId: uuid.MustParse(d.OrganisationId),
		Id:             uuid.MustParse(d.Id),
		Type:           d.Type,
		Data:           d.Data,
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

	if r.Tags != nil && len(r.Tags) != 0 {
		d.TypeTarget += separator + strings.Join(r.Tags, separator)
	}

	return d
}

func (d *dto) createEdge() Edge {
	typeTarget := strings.Split(d.TypeTarget, separator)
	edge := Edge{
		OrganisationId: uuid.MustParse(d.OrganisationId),
		Id:             uuid.MustParse(d.Id),
		TargetNodeId:   uuid.MustParse(typeTarget[1]),
		TargetNodeType: d.Type,
		Data:           d.Data,
	}
	if len(typeTarget) > 2 {
		edge.Tags = typeTarget[2:]
	}
	return edge
}
