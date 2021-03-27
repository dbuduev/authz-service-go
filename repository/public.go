package repository

import (
	"fmt"
	"github.com/dbuduev/authz-service-go/core"
	"github.com/dbuduev/authz-service-go/dygraph"
	"github.com/google/uuid"
)

const (
	RoleRecordType        = "ROLE"
	OperationRecordType   = "OP"
	BranchRecordType      = "BRANCH"
	BranchGroupRecordType = "BRANCH_GROUP"
)

type Repository struct {
	graphDB GraphDB
}

func (r *Repository) AddOperation(op core.Operation) error {
	fmt.Printf("Adding operation %v\n", op)
	return r.graphDB.InsertRecord(&dygraph.LogicalRecordRequest{
		OrganisationId: op.OrganisationId,
		Id:             op.Id,
		Type:           OperationRecordType,
		Data:           op.Name,
	})
}

func (r *Repository) AddRole(role core.Role) error {
	fmt.Printf("Adding role %v\n", role)
	return r.graphDB.InsertRecord(&dygraph.LogicalRecordRequest{
		OrganisationId: role.OrganisationId,
		Id:             role.Id,
		Type:           RoleRecordType,
		Data:           role.Name,
	})
}

func (r *Repository) AddBranch(b core.Branch) error {
	return r.graphDB.InsertRecord(&dygraph.LogicalRecordRequest{
		OrganisationId: b.OrganisationId,
		Id:             b.Id,
		Type:           BranchRecordType,
		Data:           b.Name,
	})
}

func (r *Repository) AddBranchGroup(g core.BranchGroup) error {
	return r.graphDB.InsertRecord(&dygraph.LogicalRecordRequest{
		OrganisationId: g.OrganisationId,
		Id:             g.Id,
		Type:           BranchGroupRecordType,
		Data:           g.Name,
	})
}

func (r *Repository) AssignOperationToRole(x core.OperationAssignment) error {
	fmt.Printf("Assigning operation to role %v\n", x)
	request := []dygraph.CreateEdgeRequest{
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

	return r.graphDB.TransactionalInsert(request)
}

func (r *Repository) AssignBranchToBranchGroup(x core.BranchAssignment) error {
	fmt.Printf("Assigning branch to branch group %v\n", x)
	request := []dygraph.CreateEdgeRequest{
		{
			OrganisationId: x.OrganisationId,
			Id:             x.BranchId,
			TargetNodeId:   x.BranchGroupId,
			TargetNodeType: BranchGroupRecordType,
		},
		{
			OrganisationId: x.OrganisationId,
			Id:             x.BranchGroupId,
			TargetNodeId:   x.BranchId,
			TargetNodeType: BranchRecordType,
		},
	}

	return r.graphDB.TransactionalInsert(request)
}

func (r *Repository) GetBranchesByBranchGroup(organisationId, branchGroupId uuid.UUID) ([]uuid.UUID, error) {
	items, err := r.graphDB.GetNodeEdgesOfType(organisationId, branchGroupId, BranchRecordType)
	if err != nil {
		return nil, err
	}

	result := make([]uuid.UUID, len(items))
	for i, item := range items {
		result[i] = uuid.MustParse(item.TypeTarget[1])
	}

	return result, nil
}

func (r *Repository) GetRolesByOperation(organisationId, opId uuid.UUID) ([]uuid.UUID, error) {
	items, err := r.graphDB.GetNodeEdgesOfType(organisationId, opId, RoleRecordType)
	if err != nil {
		return nil, err
	}

	result := make([]uuid.UUID, len(items))
	for i, item := range items {
		result[i] = uuid.MustParse(item.TypeTarget[1])
	}

	return result, nil
}

func (r *Repository) GetOperationsByRole(organisationId, roleId uuid.UUID) ([]uuid.UUID, error) {
	items, err := r.graphDB.GetNodeEdgesOfType(organisationId, roleId, OperationRecordType)
	if err != nil {
		return nil, err
	}

	result := make([]uuid.UUID, len(items))
	for i, item := range items {
		result[i] = uuid.MustParse(item.TypeTarget[1])
	}

	return result, nil
}

func (r *Repository) getAllRoles(organisationId uuid.UUID) ([]core.Role, error) {
	nodes, err := r.graphDB.GetNodes(organisationId, RoleRecordType)
	if err != nil {
		return nil, err
	}
	result := make([]core.Role, len(nodes))
	for i, node := range nodes {
		result[i] = node.ToRole()
	}

	return result, nil
}

//func GetTags(a core.UserRoleAssignment) []string {
//	return []string{"ASSIGNED_IN_BRANCH", a.BranchId.String()}
//}
