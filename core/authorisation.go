package core

import (
	"github.com/google/uuid"
)

type AuthorisationCore struct {
	repository Repository
}

func CreateAuthorisationCore(repository Repository) AuthorisationCore {
	return AuthorisationCore{repository: repository}
}

// FindOpByName returns nil if operation is not found.
func (ac *AuthorisationCore) FindOpByName(organisationId uuid.UUID, name string) *Operation {
	ops, err := ac.repository.GetAllOperations(organisationId)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(ops); i++ {
		if ops[i].Name == name {
			return &ops[i]
		}
	}

	return nil
}

// WhereAuthorised returns a slice of branch or branch group ids where the operation is authorised for the user.
func (ac *AuthorisationCore) WhereAuthorised(organisationId, userId, opId uuid.UUID) []uuid.UUID {
	r := ac.repository

	// 1. op -> [role]
	roles, err := r.GetRolesByOperation(organisationId, opId)
	if err != nil {
		panic(err)
	}
	// no roles supporting this operation. TODO: log with warning level.
	if len(roles) == 0 {
		return nil
	}

	// 2. uid, role -> B, where B = [b|bg]
	assignments, err := r.GetUserRolesAssignments(organisationId, userId)
	if err != nil {
		panic(err)
	}

	branches := make(map[uuid.UUID]struct{}, len(assignments))
	for _, assignment := range assignments {
		for _, role := range roles {
			if role == assignment.RoleId {
				branches[assignment.BranchId] = struct{}{}
			}
		}
	}

	if len(branches) == 0 {
		return nil
	} else {
		result := make([]uuid.UUID, 0, len(branches))
		for b := range branches {
			result = append(result, b)
		}
		return result
	}
}
