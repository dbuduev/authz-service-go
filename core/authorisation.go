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

func (ac *AuthorisationCore) WhereAuthorised(organisationId, userId uuid.UUID, operation string) []UserRoleAssignment {
	r := ac.repository

	op := ac.FindOpByName(organisationId, operation)
	if op == nil {
		panic("operation not found " + operation)
	}
	roles, err := r.GetRolesByOperation(organisationId, op.Id)
	if err != nil {
		panic(err)
	}
	if len(roles) == 0 {
		return nil
	}

	panic("not implemented")
}
