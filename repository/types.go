package repository

import (
	"github.com/dbuduev/authz-service-go/dygraph"
	"github.com/google/uuid"
)

type GraphDB interface {
	InsertRecord(node *dygraph.Node) error
	GetNodes(organisationId uuid.UUID, nodeType string) ([]dygraph.Node, error)
	GetEdges(organisationId uuid.UUID, edgeType string) ([]dygraph.Edge, error)
	GetNodeEdgesOfType(organisationId, id uuid.UUID, edgeType string) ([]dygraph.Edge, error)
	TransactionalInsert(items []dygraph.Edge) error
}
