package repository

import (
	"fmt"
	"github.com/dbuduev/authz-service-go/core"
	"github.com/dbuduev/authz-service-go/dygraph"
	"github.com/dbuduev/authz-service-go/sphinx"
	"github.com/google/uuid"
)

const (
	RoleRecordType        = "ROLE"
	OperationRecordType   = "OP"
	BranchRecordType      = "BRANCH"
	BranchGroupRecordType = "BRANCH_GROUP"
	UserRecordType        = "USER"
)

type Repository struct {
	graphDB GraphDB
}

func CreateRepository(graphDB GraphDB) *Repository {
	return &Repository{graphDB: graphDB}
}

func (r *Repository) AddOperation(op core.Operation) error {
	fmt.Printf("Adding operation %v\n", op)
	return r.graphDB.InsertRecord(&dygraph.Node{
		OrganisationId: op.OrganisationId,
		Id:             op.Id,
		Type:           OperationRecordType,
		Data:           op.Name,
	})
}

func (r *Repository) AddRole(role core.Role) error {
	fmt.Printf("Adding role %v\n", role)
	return r.graphDB.InsertRecord(&dygraph.Node{
		OrganisationId: role.OrganisationId,
		Id:             role.Id,
		Type:           RoleRecordType,
		Data:           role.Name,
	})
}

func (r *Repository) AddBranch(b core.Branch) error {
	return r.graphDB.InsertRecord(&dygraph.Node{
		OrganisationId: b.OrganisationId,
		Id:             b.Id,
		Type:           BranchRecordType,
		Data:           b.Name,
	})
}

func (r *Repository) AddBranchGroup(g core.BranchGroup) error {
	return r.graphDB.InsertRecord(&dygraph.Node{
		OrganisationId: g.OrganisationId,
		Id:             g.Id,
		Type:           BranchGroupRecordType,
		Data:           g.Name,
	})
}

func (r *Repository) AssignOperationToRole(x core.OperationAssignment) error {
	fmt.Printf("Assigning operation to role %v\n", x)
	request := []dygraph.Edge{
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
	request := []dygraph.Edge{
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
		result[i] = item.TargetNodeId
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
		result[i] = item.TargetNodeId
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
		result[i] = item.TargetNodeId
	}

	return result, nil
}

func (r *Repository) GetAllRoles(organisationId uuid.UUID) ([]core.Role, error) {
	nodes, err := r.graphDB.GetNodes(organisationId, RoleRecordType)
	if err != nil {
		return nil, err
	}
	result := make([]core.Role, len(nodes))
	for i, node := range nodes {
		result[i] = ToRole(node)
	}

	return result, nil
}

func (r *Repository) GetAllOperations(organisationId uuid.UUID) ([]core.Operation, error) {
	nodes, err := r.graphDB.GetNodes(organisationId, OperationRecordType)
	if err != nil {
		return nil, err
	}
	result := make([]core.Operation, len(nodes))
	for i, node := range nodes {
		result[i] = ToOperation(node)
	}

	return result, nil
}

func (r *Repository) AssignRoleToUser(x core.UserRoleAssignment) error {
	fmt.Printf("Assigning role to a user in a branch %v\n", x)
	tags := []string{"ASSIGNED_IN_BRANCH", x.BranchId.String()}
	request := []dygraph.Edge{
		{
			OrganisationId: x.OrganisationId,
			Id:             x.RoleId,
			TargetNodeId:   x.UserId,
			TargetNodeType: UserRecordType,
			Tags:           tags,
			Data:           x.BranchId.String(),
		},
		{
			OrganisationId: x.OrganisationId,
			Id:             x.UserId,
			TargetNodeId:   x.RoleId,
			TargetNodeType: RoleRecordType,
			Tags:           tags,
			Data:           x.BranchId.String(),
		},
	}

	return r.graphDB.TransactionalInsert(request)
}

func (r *Repository) GetUserRolesAssignments(organisationId, userId uuid.UUID) ([]core.UserRoleAssignment, error) {
	records, err := r.graphDB.GetNodeEdgesOfType(organisationId, userId, RoleRecordType)
	if err != nil {
		return nil, err
	}
	result := make([]core.UserRoleAssignment, len(records))
	for i, record := range records {
		result[i] = ToUserRoleAssignment(record)
	}

	return result, nil
}

func (r *Repository) GetHierarchy(organisationId uuid.UUID) (sphinx.BranchGroupContent, error) {
	links, err := r.graphDB.GetEdges(organisationId, BranchGroupRecordType)
	if err != nil {
		return nil, err
	}
	result := make(sphinx.BranchGroupContent, len(links))

	for _, link := range links {
		branchGroupId := link.TargetNodeId
		result[branchGroupId] = append(result[branchGroupId], link.Id)
	}
	return result, nil
}

func ToRole(r dygraph.Node) core.Role {
	return core.Role{
		OrganisationId: r.OrganisationId,
		Id:             r.Id,
		Name:           r.Data,
	}
}

func ToOperation(r dygraph.Node) core.Operation {
	return core.Operation{
		OrganisationId: r.OrganisationId,
		Id:             r.Id,
		Name:           r.Data,
	}
}

func ToUserRoleAssignment(r dygraph.Edge) core.UserRoleAssignment {
	return core.UserRoleAssignment{
		OrganisationId: r.OrganisationId,
		RoleId:         r.TargetNodeId,
		UserId:         r.Id,
		BranchId:       uuid.MustParse(r.Data),
	}
}
