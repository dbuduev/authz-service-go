package core

import (
	"github.com/dbuduev/authz-service-go/sphinx"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"reflect"
	"sort"
	"testing"
)

type testRepository struct {
	addOperation              func(op Operation) error
	addRole                   func(role Role) error
	addBranch                 func(b Branch) error
	addBranchGroup            func(g BranchGroup) error
	assignOperationToRole     func(x OperationAssignment) error
	assignBranchToBranchGroup func(x BranchAssignment) error
	getBranchesByBranchGroup  func(organisationId, branchGroupId uuid.UUID) ([]uuid.UUID, error)
	getRolesByOperation       func(organisationId, opId uuid.UUID) ([]uuid.UUID, error)
	getOperationsByRole       func(organisationId, roleId uuid.UUID) ([]uuid.UUID, error)
	getAllRoles               func(organisationId uuid.UUID) ([]Role, error)
	getAllOperations          func(organisationId uuid.UUID) ([]Operation, error)
	assignRoleToUser          func(x UserRoleAssignment) error
	getUserRolesAssignments   func(organisationId, userId uuid.UUID) ([]UserRoleAssignment, error)
	getHierarchy              func(organisationId uuid.UUID) (sphinx.BranchGroupContent, error)
}

func (t testRepository) AddOperation(op Operation) error {
	return t.addOperation(op)
}

func (t testRepository) AddRole(role Role) error {
	return t.addRole(role)
}

func (t testRepository) AddBranch(b Branch) error {
	return t.addBranch(b)
}

func (t testRepository) AddBranchGroup(g BranchGroup) error {
	return t.addBranchGroup(g)
}

func (t testRepository) AssignOperationToRole(x OperationAssignment) error {
	return t.assignOperationToRole(x)
}

func (t testRepository) AssignBranchToBranchGroup(x BranchAssignment) error {
	return t.assignBranchToBranchGroup(x)
}

func (t testRepository) GetBranchesByBranchGroup(organisationId, branchGroupId uuid.UUID) ([]uuid.UUID, error) {
	return t.getBranchesByBranchGroup(organisationId, branchGroupId)
}

func (t testRepository) GetRolesByOperation(organisationId, opId uuid.UUID) ([]uuid.UUID, error) {
	return t.getRolesByOperation(organisationId, opId)
}

func (t testRepository) GetOperationsByRole(organisationId, roleId uuid.UUID) ([]uuid.UUID, error) {
	return t.getOperationsByRole(organisationId, roleId)
}

func (t testRepository) GetAllRoles(organisationId uuid.UUID) ([]Role, error) {
	return t.getAllRoles(organisationId)
}

func (t testRepository) GetAllOperations(organisationId uuid.UUID) ([]Operation, error) {
	return t.getAllOperations(organisationId)
}

func (t testRepository) AssignRoleToUser(x UserRoleAssignment) error {
	return t.assignRoleToUser(x)
}

func (t testRepository) GetUserRolesAssignments(organisationId, userId uuid.UUID) ([]UserRoleAssignment, error) {
	return t.getUserRolesAssignments(organisationId, userId)
}

func (t testRepository) GetHierarchy(organisationId uuid.UUID) (sphinx.BranchGroupContent, error) {
	return t.getHierarchy(organisationId)
}

func GenId(id uuid.UUID, b byte) uuid.UUID {
	return uuid.NewSHA1(id, []byte{b})
}

func TestAuthorisationCore_FindOpByName(t *testing.T) {
	repository := testRepository{}
	ac := &AuthorisationCore{
		repository: &repository,
	}

	type args struct {
		organisationId uuid.UUID
		name           string
	}
	tests := []struct {
		name             string
		getAllOperations func(orgId uuid.UUID) ([]Operation, error)
		args             args
		want             func(orgId uuid.UUID) *Operation
	}{
		{
			name: "There's a match",
			getAllOperations: func(orgId uuid.UUID) ([]Operation, error) {
				return []Operation{{orgId, GenId(orgId, 1), "manage-staff"}, {orgId, GenId(orgId, 2), "view-staff"}}, nil
			},
			args: args{
				organisationId: uuid.New(),
				name:           "view-staff",
			},
			want: func(orgId uuid.UUID) *Operation {
				return &Operation{
					OrganisationId: orgId,
					Id:             GenId(orgId, 2),
					Name:           "view-staff",
				}
			},
		},
		{
			name: "There's no match",
			getAllOperations: func(orgId uuid.UUID) ([]Operation, error) {
				return []Operation{{orgId, GenId(orgId, 1), "manage-staff"}, {orgId, GenId(orgId, 2), "view-staff"}}, nil
			},
			args: args{
				organisationId: uuid.New(),
				name:           "no-such-thing",
			},
			want: func(orgId uuid.UUID) *Operation {
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository.getAllOperations = tt.getAllOperations
			want := tt.want(tt.args.organisationId)
			got := ac.FindOpByName(tt.args.organisationId, tt.args.name)
			if want == nil {
				if got != nil {
					t.Errorf("FindOpByName() = %v, want nil", *got)
				}
			} else if got == nil || !reflect.DeepEqual(*got, *want) {
				t.Errorf("FindOpByName() = %v, want %v", got, want)
			}
		})
	}
}

func TestAuthorisationCore_WhereAuthorised(t *testing.T) {
	trans := cmp.Transformer("Sort", func(in []uuid.UUID) []uuid.UUID {
		out := append([]uuid.UUID(nil), in...) // Copy input to avoid mutating it
		sort.Slice(out, func(i, j int) bool {
			for k := 0; k < len(out[i]); k++ {
				if out[i][k] != out[j][k] {
					return out[i][k] < out[j][k]
				}
			}
			return false
		})
		return out
	})

	repository := testRepository{}
	ac := &AuthorisationCore{
		repository: &repository,
	}

	type args struct {
		organisationId uuid.UUID
		userId         byte
		opId           byte
	}
	tests := []struct {
		name                    string
		args                    args
		getRolesByOperation     func(orgId, opId uuid.UUID) ([]uuid.UUID, error)
		getUserRolesAssignments func(orgId, userId uuid.UUID) ([]UserRoleAssignment, error)
		want                    func(orgId uuid.UUID) []uuid.UUID
	}{
		{
			name: "No assignments",
			args: args{
				organisationId: uuid.New(),
				userId:         1,
				opId:           2,
			},
			getRolesByOperation:     func(_, _ uuid.UUID) ([]uuid.UUID, error) { return nil, nil },
			getUserRolesAssignments: func(_, _ uuid.UUID) ([]UserRoleAssignment, error) { return nil, nil },
			want:                    func(_ uuid.UUID) []uuid.UUID { return nil },
		},
		{
			name: "A single assignment",
			args: args{
				organisationId: uuid.New(),
				userId:         1,
				opId:           2,
			},
			getRolesByOperation: func(orgId, _ uuid.UUID) ([]uuid.UUID, error) {
				return []uuid.UUID{GenId(orgId, 1), GenId(orgId, 2)}, nil
			},
			getUserRolesAssignments: func(orgId, userId uuid.UUID) ([]UserRoleAssignment, error) {
				return []UserRoleAssignment{
					{
						OrganisationId: orgId,
						RoleId:         GenId(orgId, 2),
						UserId:         userId,
						BranchId:       GenId(orgId, 10),
					},
				}, nil
			},
			want: func(orgId uuid.UUID) []uuid.UUID {
				return []uuid.UUID{GenId(orgId, 10)}
			},
		},
		{
			name: "Two assignments for the same branch",
			args: args{
				organisationId: uuid.New(),
				userId:         1,
				opId:           2,
			},
			getRolesByOperation: func(orgId, _ uuid.UUID) ([]uuid.UUID, error) {
				return []uuid.UUID{GenId(orgId, 1), GenId(orgId, 2)}, nil
			},
			getUserRolesAssignments: func(orgId, userId uuid.UUID) ([]UserRoleAssignment, error) {
				return []UserRoleAssignment{
					{
						OrganisationId: orgId,
						RoleId:         GenId(orgId, 2),
						UserId:         userId,
						BranchId:       GenId(orgId, 10),
					},
					{
						OrganisationId: orgId,
						RoleId:         GenId(orgId, 1),
						UserId:         userId,
						BranchId:       GenId(orgId, 10),
					},
				}, nil
			},
			want: func(orgId uuid.UUID) []uuid.UUID {
				return []uuid.UUID{GenId(orgId, 10)}
			},
		},
		{
			name: "Two assignments",
			args: args{
				organisationId: uuid.New(),
				userId:         1,
				opId:           2,
			},
			getRolesByOperation: func(orgId, _ uuid.UUID) ([]uuid.UUID, error) {
				return []uuid.UUID{GenId(orgId, 1), GenId(orgId, 2)}, nil
			},
			getUserRolesAssignments: func(orgId, userId uuid.UUID) ([]UserRoleAssignment, error) {
				return []UserRoleAssignment{
					{
						OrganisationId: orgId,
						RoleId:         GenId(orgId, 2),
						UserId:         userId,
						BranchId:       GenId(orgId, 10),
					},
					{
						OrganisationId: orgId,
						RoleId:         GenId(orgId, 1),
						UserId:         userId,
						BranchId:       GenId(orgId, 11),
					},
				}, nil
			},
			want: func(orgId uuid.UUID) []uuid.UUID {
				return []uuid.UUID{GenId(orgId, 10), GenId(orgId, 11)}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository.getRolesByOperation = tt.getRolesByOperation
			repository.getUserRolesAssignments = tt.getUserRolesAssignments

			want := tt.want(tt.args.organisationId)
			got := ac.WhereAuthorised(tt.args.organisationId, GenId(tt.args.organisationId, tt.args.userId), GenId(tt.args.organisationId, tt.args.opId))
			if diff := cmp.Diff(want, got, trans); diff != "" {
				t.Errorf("WhereAuthorised() diff  %v", diff)
			}
		})
	}
}
