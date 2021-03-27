package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/dbuduev/authz-service-go/core"
	"github.com/dbuduev/authz-service-go/dygraph"
	"github.com/google/uuid"
	"reflect"
	"sort"
	"testing"
)

type Operation struct {
	OrganisationId byte
	Id             byte
	Name           string
}

func (o Operation) To(id uuid.UUID) core.Operation {
	return core.Operation{
		OrganisationId: GenId(id, o.OrganisationId),
		Id:             GenId(id, o.Id),
		Name:           o.Name,
	}
}

type Role struct {
	OrganisationId byte
	Id             byte
	Name           string
}

func GenId(id uuid.UUID, b byte) uuid.UUID {
	return uuid.NewSHA1(id, []byte{b})
}

func (r Role) To(id uuid.UUID) core.Role {
	return core.Role{
		OrganisationId: GenId(id, r.OrganisationId),
		Id:             GenId(id, r.Id),
		Name:           r.Name,
	}
}

type OperationAssignment struct {
	OrganisationId byte
	RoleId         byte
	OperationId    byte
}

func (a OperationAssignment) To(id uuid.UUID) core.OperationAssignment {
	return core.OperationAssignment{
		OrganisationId: GenId(id, a.OrganisationId),
		RoleId:         GenId(id, a.RoleId),
		OperationId:    GenId(id, a.OperationId),
	}
}

type Branch struct {
	OrganisationId byte
	Id             byte
	Name           string
}

func (b Branch) To(id uuid.UUID) core.Branch {
	return core.Branch{
		OrganisationId: GenId(id, b.OrganisationId),
		Id:             GenId(id, b.Id),
		Name:           b.Name,
	}
}

type BranchGroup struct {
	OrganisationId byte
	Id             byte
	Name           string
}

func (b BranchGroup) To(id uuid.UUID) core.BranchGroup {
	return core.BranchGroup{
		OrganisationId: GenId(id, b.OrganisationId),
		Id:             GenId(id, b.Id),
		Name:           b.Name,
	}
}

type BranchAssignment struct {
	OrganisationId byte
	BranchId       byte
	BranchGroupId  byte
}

func (a BranchAssignment) To(id uuid.UUID) core.BranchAssignment {
	return core.BranchAssignment{
		OrganisationId: GenId(id, a.OrganisationId),
		BranchId:       GenId(id, a.BranchId),
		BranchGroupId:  GenId(id, a.BranchGroupId),
	}
}

type testConfig struct {
	roles             []Role
	operations        []Operation
	assignments       []OperationAssignment
	branches          []Branch
	branchGroups      []BranchGroup
	branchAssignments []BranchAssignment
}

func setUpTest(repository Repository, config testConfig, id uuid.UUID) {
	for _, r := range config.roles {
		err := repository.AddRole(r.To(id))
		if err != nil {
			panic(err)
		}
	}
	for _, operation := range config.operations {
		err := repository.AddOperation(operation.To(id))
		if err != nil {
			panic(err)
		}

	}
	for _, assignment := range config.assignments {
		err := repository.AssignOperationToRole(assignment.To(id))
		if err != nil {
			panic(err)
		}
	}

	for _, b := range config.branches {
		if err := repository.AddBranch(b.To(id)); err != nil {
			panic(err)
		}
	}

	for _, b := range config.branchGroups {
		if err := repository.AddBranchGroup(b.To(id)); err != nil {
			panic(err)
		}
	}

	for _, b := range config.branchAssignments {
		if err := repository.AssignBranchToBranchGroup(b.To(id)); err != nil {
			panic(err)
		}
	}
}

func TestRepository_GetRolesByOperation(t *testing.T) {
	repository := CreateTestRepository()

	type args struct {
		organisationId byte
		opId           byte
	}

	tests := []struct {
		id      uuid.UUID
		name    string
		config  testConfig
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "Empty",
			id:      uuid.New(),
			config:  testConfig{},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "Simple case",
			id:   uuid.New(),
			config: testConfig{
				roles:       []Role{{1, 3, "Admin"}, {1, 4, "PT"}},
				operations:  []Operation{{1, 2, "manage-member"}, {1, 5, "view-member"}},
				assignments: []OperationAssignment{{1, 3, 2}, {1, 3, 5}, {1, 4, 5}},
			},
			args: args{
				organisationId: 1,
				opId:           2,
			},
			want:    []byte{3},
			wantErr: false,
		},
		{
			name: "Two roles having the operation",
			id:   uuid.New(),
			config: testConfig{
				roles:       []Role{{1, 3, "Admin"}, {1, 4, "Owner"}},
				operations:  []Operation{{1, 2, "manage-member"}, {1, 5, "view-member"}},
				assignments: []OperationAssignment{{1, 3, 2}, {1, 4, 2}, {1, 4, 5}},
			},
			args: args{
				organisationId: 1,
				opId:           2,
			},
			want:    []byte{3, 4},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setUpTest(repository, tt.config, tt.id)
			got, err := repository.GetRolesByOperation(GenId(tt.id, tt.args.organisationId), GenId(tt.id, tt.args.opId))
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRolesByOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := make([]uuid.UUID, len(tt.want))
			for i, roleId := range tt.want {
				want[i] = GenId(tt.id, roleId)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("GetRolesByOperation() got = %v, want %v", got, want)
			}
		})
	}
}
func TestRepository_GetOperationsByRole(t *testing.T) {
	repository := CreateTestRepository()

	type args struct {
		organisationId byte
		roleId         byte
	}

	tests := []struct {
		id      uuid.UUID
		name    string
		config  testConfig
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "Empty",
			id:      uuid.New(),
			config:  testConfig{},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "Simple case",
			id:   uuid.New(),
			config: testConfig{
				roles:       []Role{{1, 3, "Admin"}, {1, 4, "PT"}},
				operations:  []Operation{{1, 2, "manage-member"}, {1, 5, "view-member"}},
				assignments: []OperationAssignment{{1, 3, 2}, {1, 3, 5}, {1, 4, 5}},
			},
			args: args{
				organisationId: 1,
				roleId:         3,
			},
			want:    []byte{2, 5},
			wantErr: false,
		},
		{
			name: "Just a single operation",
			id:   uuid.New(),
			config: testConfig{
				roles:       []Role{{1, 3, "Admin"}, {1, 4, "Owner"}},
				operations:  []Operation{{1, 2, "manage-member"}, {1, 5, "view-member"}},
				assignments: []OperationAssignment{{1, 3, 2}, {1, 4, 2}, {1, 4, 5}},
			},
			args: args{
				organisationId: 1,
				roleId:         3,
			},
			want:    []byte{2},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setUpTest(repository, tt.config, tt.id)
			got, err := repository.GetOperationsByRole(GenId(tt.id, tt.args.organisationId), GenId(tt.id, tt.args.roleId))
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOperationsByRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := make([]uuid.UUID, len(tt.want))
			for i, roleId := range tt.want {
				want[i] = GenId(tt.id, roleId)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("GetRolesByOperation() got = %v, want %v", got, want)
			}
		})
	}
}

func TestRepository_getAllRoles(t *testing.T) {
	repository := CreateTestRepository()

	type args struct {
		organisationId byte
	}
	tests := []struct {
		name    string
		id      uuid.UUID
		config  testConfig
		args    args
		want    []Role
		wantErr bool
	}{
		{
			name:    "Empty",
			id:      uuid.New(),
			config:  testConfig{},
			want:    []Role{},
			wantErr: false,
		},
		{
			name: "2 roles",
			id:   uuid.New(),
			config: testConfig{
				roles: []Role{{1, 3, "Admin"}, {1, 4, "PT"}},
			},
			args: args{
				organisationId: 1,
			},
			wantErr: false,
		},
		{
			name: "3 roles, different organisations",
			id:   uuid.New(),
			config: testConfig{
				roles: []Role{{1, 3, "Admin"}, {2, 5, "Admin"}, {1, 4, "PT"}},
			},
			args: args{
				organisationId: 1,
			},
			want:    []Role{{1, 3, "Admin"}, {1, 4, "PT"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setUpTest(repository, tt.config, tt.id)
			organisationId := GenId(tt.id, tt.args.organisationId)
			got, err := repository.getAllRoles(organisationId)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAllRoles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var want []core.Role
			copyFrom := tt.want
			if copyFrom == nil {
				copyFrom = tt.config.roles
			}
			want = make([]core.Role, len(copyFrom))
			for i, role := range copyFrom {
				want[i] = role.To(tt.id)
			}
			sort.Slice(want, func(i, j int) bool {
				return want[i].Name < want[j].Name
			})
			sort.Slice(got, func(i, j int) bool {
				return got[i].Name < got[j].Name
			})
			if !reflect.DeepEqual(got, want) {
				t.Errorf("getAllRoles() got = %v, want %v", got, want)
			}
		})
	}
}

func TestRepository_GetBranchesByBranchGroup(t *testing.T) {
	repository := CreateTestRepository()

	type args struct {
		organisationId byte
		branchGroupId  byte
	}

	tests := []struct {
		id      uuid.UUID
		name    string
		config  testConfig
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "Empty",
			id:      uuid.New(),
			config:  testConfig{},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "Simple case",
			id:   uuid.New(),
			config: testConfig{
				branches:          []Branch{{1, 3, "A"}, {1, 4, "B"}},
				branchGroups:      []BranchGroup{{1, 2, "X"}, {1, 5, "Y"}},
				branchAssignments: []BranchAssignment{{1, 3, 2}, {1, 3, 5}, {1, 4, 5}},
			},
			args: args{
				organisationId: 1,
				branchGroupId:  5,
			},
			want:    []byte{3, 4},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setUpTest(repository, tt.config, tt.id)
			got, err := repository.GetBranchesByBranchGroup(GenId(tt.id, tt.args.organisationId), GenId(tt.id, tt.args.branchGroupId))
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBranchesByBranchGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := make([]uuid.UUID, len(tt.want))
			for i, branchId := range tt.want {
				want[i] = GenId(tt.id, branchId)
			}
			sort.Slice(want, func(i, j int) bool {
				return want[i][0] < want[j][0] // sort UUIDs just by the first byte.
			})
			sort.Slice(got, func(i, j int) bool {
				return got[i][0] < got[j][0]
			})

			if !reflect.DeepEqual(got, want) {
				t.Errorf("GetBranchesByBranchGroup() got = %v, want %v", got, want)
			}
		})
	}
}
func GetClient() *dynamodb.DynamoDB {
	// Create DynamoDB client
	s := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return dynamodb.New(s, s.Config.WithEndpoint("http://localhost:8000"))
}

func CreateTestGraphClient() *dygraph.Dygraph {
	return dygraph.CreateGraphClient(GetClient(), "test")
}

func CreateTestRepository() Repository {
	return Repository{
		graphDB: CreateTestGraphClient(),
	}
}
