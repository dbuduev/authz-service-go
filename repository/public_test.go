package repository

import (
	"github.com/dbuduev/authz-service-go/core"
	"github.com/google/uuid"
	"reflect"
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

func TestRepository_GetRolesByOperation(t *testing.T) {
	repository := CreateTestRepository()
	type testConfig struct {
		roles       []Role
		operations  []Operation
		assignments []OperationAssignment
	}
	setUpTest := func(config testConfig, id uuid.UUID) {
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
	}

	type args struct {
		organisationId byte
		opId           byte
	}

	tests := []struct {
		id      uuid.UUID
		name    string
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
			name: "Simple case",
			id:   uuid.New(),
			config: testConfig{
				roles:       []Role{{1, 3, "Admin"}, {1, 4, "PT"}},
				operations:  []Operation{{OrganisationId: 1, Id: 2, Name: "manage-member"}, {OrganisationId: 1, Id: 5, Name: "view-member"}},
				assignments: []OperationAssignment{{OrganisationId: 1, RoleId: 3, OperationId: 2}, {OrganisationId: 1, RoleId: 3, OperationId: 5}, {OrganisationId: 1, RoleId: 4, OperationId: 5}},
			},
			args: args{
				organisationId: 1,
				opId:           2,
			},
			want:    []Role{{OrganisationId: 1, Id: 3, Name: "Admin"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setUpTest(tt.config, tt.id)
			got, err := repository.GetRolesByOperation(GenId(tt.id, tt.args.organisationId), GenId(tt.id, tt.args.opId))
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRolesByOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := make([]core.Role, len(tt.want))
			for i, role := range tt.want {
				want[i] = role.To(tt.id)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("GetRolesByOperation() got = %v, want %v", got, want)
			}
		})
	}
}
