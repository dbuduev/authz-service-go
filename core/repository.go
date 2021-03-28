package core

import (
	"github.com/dbuduev/authz-service-go/sphinx"
	"github.com/google/uuid"
)

type Repository interface {
	AddOperation(op Operation) error
	AddRole(role Role) error
	AddBranch(b Branch) error
	AddBranchGroup(g BranchGroup) error
	AssignOperationToRole(x OperationAssignment) error
	AssignBranchToBranchGroup(x BranchAssignment) error
	GetBranchesByBranchGroup(organisationId, branchGroupId uuid.UUID) ([]uuid.UUID, error)
	GetRolesByOperation(organisationId, opId uuid.UUID) ([]uuid.UUID, error)
	GetOperationsByRole(organisationId, roleId uuid.UUID) ([]uuid.UUID, error)
	GetAllRoles(organisationId uuid.UUID) ([]Role, error)
	AssignRoleToUser(x UserRoleAssignment) error
	GetUserRolesAssignments(organisationId, userId uuid.UUID) ([]UserRoleAssignment, error)
	GetHierarchy(organisationId uuid.UUID) (sphinx.BranchGroupContent, error)
}
