package repository

import (
	"fmt"
	"github.com/dbuduev/authz-service-go/core"
	"github.com/google/uuid"
	"strings"
)

type LogicalRecordRequest struct {
	OrganisationId uuid.UUID
	Id             uuid.UUID
	Type           string
	Data           string
}

type LogicalRecord struct {
	LogicalRecordRequest
	TypeTarget []string
}

type CreateEdgeRequest struct {
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

func (node *LogicalRecordRequest) createNodeDto() *dto {
	return &dto{
		GlobalId:       fmt.Sprintf("%s_%s", node.OrganisationId, node.Id),
		TypeTarget:     nodePrefix + strings.Join([]string{node.Type, node.Id.String()}, separator),
		OrganisationId: node.OrganisationId.String(),
		Id:             node.Id.String(),
		Type:           node.Type,
		Data:           node.Data,
	}
}

func (r *CreateEdgeRequest) createNodeDto() *dto {
	d := &dto{
		GlobalId:       fmt.Sprintf("%s_%s", r.OrganisationId, r.Id),
		TypeTarget:     edgePrefix + r.TargetNodeType + separator + r.TargetNodeId.String(),
		OrganisationId: r.OrganisationId.String(),
		Id:             r.Id.String(),
		Type:           r.TargetNodeType,
		Data:           r.Data,
	}
	if d.Data == "" {
		d.Data = r.TargetNodeId.String()
	}

	return d
}

func (r LogicalRecord) ToRole() core.Role {
	return core.Role{
		OrganisationId: r.OrganisationId,
		Id:             r.Id,
		Name:           r.Data,
	}
}
