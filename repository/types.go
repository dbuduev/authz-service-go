package repository

import (
	"github.com/dbuduev/authz-service-go/dygraph"
	"github.com/google/uuid"
)

type GraphDB interface {
	InsertRecord(node *dygraph.LogicalRecordRequest) error
	GetNodes(organisationId uuid.UUID, nodeType string) ([]dygraph.LogicalRecord, error)
	GetEdges(organisationId uuid.UUID, edgeType string) ([]dygraph.LogicalRecord, error)
	GetNodeEdgesOfType(organisationId, id uuid.UUID, edgeType string) ([]dygraph.LogicalRecord, error)
	TransactionalInsert(items []dygraph.CreateEdgeRequest) error
}
