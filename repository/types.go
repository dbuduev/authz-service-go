package repository

import "github.com/google/uuid"

type Node struct {
	OrganisationId uuid.UUID
	Id             uuid.UUID
	Type           string
	Data           string
}
