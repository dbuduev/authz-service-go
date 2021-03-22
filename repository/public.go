package repository

import (
	"fmt"
	"github.com/dbuduev/authz-service-go/core"
	"github.com/google/uuid"
)

const (
	RoleRecordType        = "ROLE"
	OperationRecordType   = "OP"
	BranchRecordType      = "BRANCH"
	BranchGroupRecordType = "BRANCH_GROUP"
)

func (r *Repository) AddOperation(op core.Operation) error {
	fmt.Printf("Adding operation %v\n", op)
	return r.insertRecord(&LogicalRecordRequest{
		OrganisationId: op.OrganisationId,
		Id:             op.Id,
		Type:           OperationRecordType,
		Data:           op.Name,
	})
}

func (r *Repository) AddRole(role core.Role) error {
	fmt.Printf("Adding role %v\n", role)
	return r.insertRecord(&LogicalRecordRequest{
		OrganisationId: role.OrganisationId,
		Id:             role.Id,
		Type:           RoleRecordType,
		Data:           role.Name,
	})
}

func (r *Repository) AddBranch(b core.Branch) error {
	return r.insertRecord(&LogicalRecordRequest{
		OrganisationId: b.OrganisationId,
		Id:             b.Id,
		Type:           BranchRecordType,
		Data:           b.Name,
	})
}

func (r *Repository) AddBranchGroup(g core.BranchGroup) error {
	return r.insertRecord(&LogicalRecordRequest{
		OrganisationId: g.OrganisationId,
		Id:             g.Id,
		Type:           BranchGroupRecordType,
		Data:           g.Name,
	})
}

func (r *Repository) AssignOperationToRole(x core.OperationAssignment) error {
	fmt.Printf("Assigning operation to role %v\n", x)
	request := []CreateEdgeRequest{
		{
			OrganisationId: x.OrganisationId,
			Id:             x.OperationId,
			TargetNodeId:   x.RoleId,
			TargetNodeType: RoleRecordType,
		},
		{
			OrganisationId: x.OrganisationId,
			Id:             x.RoleId,
			TargetNodeId:   x.OperationId,
			TargetNodeType: OperationRecordType,
		},
	}

	return r.transactionalInsert(request)
}

func (r *Repository) GetRolesByOperation(organisationId, opId uuid.UUID) ([]uuid.UUID, error) {
	items, err := r.getNodeEdgesOfType(organisationId, opId, RoleRecordType)
	if err != nil {
		return nil, err
	}

	result := make([]uuid.UUID, len(items))
	for i, item := range items {
		result[i] = uuid.MustParse(item.TypeTarget[1])
	}

	return result, nil
}

//func GetTags(a core.UserRoleAssignment) []string {
//	return []string{"ASSIGNED_IN_BRANCH", a.BranchId.String()}
//}
