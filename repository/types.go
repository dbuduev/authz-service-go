package repository

import (
	"github.com/google/uuid"
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
