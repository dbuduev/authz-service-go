package core

import (
	"github.com/google/uuid"
	"sort"
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

	index := sort.Search(len(ops), func(i int) bool {
		return ops[i].Name == name
	})

	if index > 0 {
		return &ops[index]
	} else {
		return nil
	}
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
