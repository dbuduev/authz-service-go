package core

import "github.com/google/uuid"

type Operation struct {
	OrganisationId uuid.UUID
	Id             uuid.UUID
	Name           string
}

type Role struct {
	OrganisationId uuid.UUID
	Id             uuid.UUID
	Name           string
	//Description    string
	//RoleType       string
}

type Branch struct {
	OrganisationId uuid.UUID
	Id             uuid.UUID
	Name           string
}

type BranchGroup struct {
	OrganisationId uuid.UUID
	Id             uuid.UUID
	Name           string
}

type OperationAssignment struct {
	OrganisationId uuid.UUID
	RoleId         uuid.UUID
	OperationId    uuid.UUID
}

type UserRoleAssignment struct {
	OrganisationId uuid.UUID
	RoleId         uuid.UUID
	UserId         uuid.UUID
	BranchId       uuid.UUID
}
