package dygraph

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/dbuduev/authz-service-go/testutils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"reflect"
	"sort"
	"testing"
)

func TestGetNodes(t *testing.T) {
	graphClient := CreateTestGraphClient()
	node := Node{
		OrganisationId: uuid.New(),
		Id:             uuid.New(),
		Type:           "ROLE",
		Data:           "Branch manager",
	}
	err := graphClient.InsertRecord(&node)
	if err != nil {
		t.Fatalf("Failed to insert node %v with error %v", node, err)
	}

	result, err := graphClient.GetNodes(node.OrganisationId, node.Type)
	if err != nil {
		t.Fatalf("Failed to get nodes with error %v", err)
	}

	if !reflect.DeepEqual(node, result[0]) {
		t.Errorf("Expect %s, got %s", node, result[0])
	}
}

func GenId(id uuid.UUID, b byte) uuid.UUID {
	return uuid.NewSHA1(id, []byte{b})
}

func TestDygraph_MarshalErrors(t *testing.T) {
	graphClient := CreateTestGraphClient()
	graphClient.marshal = func(_ interface{}) (map[string]types.AttributeValue, error) {
		return nil, errors.New("something went wrong")
	}

	tests := []struct {
		name string
		f    func() error
	}{
		{
			name: "InsertRecord",
			f: func() error {
				return graphClient.InsertRecord(&Node{})
			},
		},
		{
			name: "TransactionalInsert",
			f: func() error {
				return graphClient.TransactionalInsert([]Edge{{}})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f()
			if err == nil {
				t.Errorf("expected an error")
			}
		})
	}
}
func TestDygraph_DuplicateErrors(t *testing.T) {
	graphClient := CreateTestGraphClient()

	tests := []struct {
		name string
		f    func() error
	}{
		{
			name: "InsertRecord",
			f: func() error {
				node := Node{
					OrganisationId: uuid.New(),
					Id:             uuid.New(),
					Type:           "PHONY",
					Data:           "PHONY",
				}
				graphClient.InsertRecord(&node)
				return graphClient.InsertRecord(&node)
			},
		},
		{
			name: "TransactionalInsert",
			f: func() error {
				edge := Edge{
					OrganisationId: uuid.New(),
					Id:             uuid.New(),
					TargetNodeId:   uuid.New(),
					TargetNodeType: "PHONY",
					Tags:           nil,
					Data:           "PHONY",
				}
				graphClient.TransactionalInsert([]Edge{edge})
				return graphClient.TransactionalInsert([]Edge{edge})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f()
			if !errors.Is(err, DuplicateError) {
				t.Errorf("expected a duplicate error")
			}
		})
	}
}

func TestDygraph_TooManyRequestsErrors(t *testing.T) {
	stub := dynamodbAPIStub{
		query: func(_ context.Context, _ *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return nil, &types.ProvisionedThroughputExceededException{}
		},
	}
	graphClient := CreateGraphClient(&stub, "test")

	tests := []struct {
		name string
		f    func() error
	}{
		{
			name: "GetNodes",
			f: func() error {
				_, err := graphClient.GetNodes(uuid.New(), "ROLE")
				return err
			},
		},
		{
			name: "GetEdges",
			f: func() error {
				_, err := graphClient.GetEdges(uuid.New(), "ROLE")
				return err
			},
		},
		{
			name: "GetNodeEdgesOfType",
			f: func() error {
				_, err := graphClient.GetNodeEdgesOfType(uuid.New(), uuid.New(), "ROLE")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f()
			if !errors.Is(err, TooManyRequestsError) {
				t.Errorf("expected too many request exception, got %v", err)
			}
		})
	}
}
func TestDygraph_UnmarshalErrors(t *testing.T) {
	stub := dynamodbAPIStub{
		query: func(_ context.Context, _ *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{
				Count: 2,
				Items: []map[string]types.AttributeValue{nil, nil},
			}, nil
		},
	}
	graphClient := CreateGraphClient(&stub, "test")
	graphClient.unmarshal = func(_ map[string]types.AttributeValue, _ interface{}) error {
		return errors.New("something went wrong")
	}

	tests := []struct {
		name string
		f    func() error
	}{
		{
			name: "GetNodes",
			f: func() error {
				_, err := graphClient.GetNodes(uuid.New(), "PHONY")
				return err
			},
		},
		{
			name: "GetEdges",
			f: func() error {
				_, err := graphClient.GetEdges(uuid.New(), "PHONY")
				return err
			},
		},
		{
			name: "GetNodeEdgesOfType",
			f: func() error {
				_, err := graphClient.GetNodeEdgesOfType(uuid.New(), uuid.New(), "PHONY")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f()
			if err == nil {
				t.Errorf("expected an error")
			}
		})
	}
}

func TestDygraph_TransactionalInsertGetEdges(t *testing.T) {
	graphClient := CreateTestGraphClient()

	type args struct {
		getItems func(id uuid.UUID) []Edge
	}
	tests := []struct {
		name string
		id   uuid.UUID
		args args
	}{
		{
			name: "Insert then retrieve two edges",
			id:   uuid.New(),
			args: args{
				getItems: func(id uuid.UUID) []Edge {
					return []Edge{
						{
							OrganisationId: id,
							Id:             GenId(id, 1),
							TargetNodeId:   GenId(id, 2),
							TargetNodeType: "ROLE",
							Tags:           []string{"tag1", "tag2"},
							Data:           "data1",
						},
						{
							OrganisationId: id,
							Id:             GenId(id, 3),
							TargetNodeId:   GenId(id, 4),
							TargetNodeType: "ROLE",
							Tags:           []string{"tag3", "tag4"},
							Data:           "data2",
						},
					}
				},
			},
		},
	}
	trans := cmp.Transformer("Sort", func(in []Edge) []Edge {
		out := append([]Edge(nil), in...) // Copy input to avoid mutating it
		sort.Slice(out, func(i, j int) bool {
			return out[i].Data < out[j].Data
		})
		return out
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edges := tt.args.getItems(tt.id)
			err := graphClient.TransactionalInsert(edges)
			if err != nil {
				t.Fatalf("Failed to insert edges %v with the error %v.", edges, err)
			}

			got, err := graphClient.GetEdges(tt.id, "ROLE")
			if err != nil {
				t.Fatalf("Failed to get edges. The error %v.", err)
			}

			if diff := cmp.Diff(edges, got, trans); diff != "" {
				t.Errorf("TransactionalInsert() vs GetEdges() diff %v", diff)
			}
		})
	}
}

func TestDygraph_TransactionalInsertGetNodeEdgesOfType(t *testing.T) {
	graphClient := CreateTestGraphClient()

	type args struct {
		getItems func(orgId, id uuid.UUID) []Edge
	}
	tests := []struct {
		name  string
		orgId uuid.UUID
		id    uuid.UUID
		args  args
	}{
		{
			name:  "Insert then retrieve two edges",
			orgId: uuid.New(),
			id:    uuid.New(),
			args: args{
				getItems: func(orgId, id uuid.UUID) []Edge {
					return []Edge{
						{
							OrganisationId: orgId,
							Id:             id,
							TargetNodeId:   GenId(id, 1),
							TargetNodeType: "ROLE",
							Tags:           []string{"tag1", "tag2"},
							Data:           "data1",
						},
						{
							OrganisationId: orgId,
							Id:             id,
							TargetNodeId:   GenId(id, 3),
							TargetNodeType: "ROLE",
							Tags:           []string{"tag3", "tag4"},
							Data:           "data2",
						},
					}
				},
			},
		},
	}
	trans := cmp.Transformer("Sort", func(in []Edge) []Edge {
		out := append([]Edge(nil), in...) // Copy input to avoid mutating it
		sort.Slice(out, func(i, j int) bool {
			return out[i].Data < out[j].Data
		})
		return out
	})

	for _, tt := range tests {
		edges := tt.args.getItems(tt.orgId, tt.id)
		err := graphClient.TransactionalInsert(edges)
		if err != nil {
			t.Fatalf("Failed to insert edges %v with the error %v.", edges, err)
		}

		got, err := graphClient.GetNodeEdgesOfType(tt.orgId, tt.id, "ROLE")
		if err != nil {
			t.Fatalf("Failed to get edges. The error %v.", err)
		}

		if diff := cmp.Diff(edges, got, trans); diff != "" {
			t.Errorf("TransactionalInsert() vs GetEdges() diff %v", diff)
		}
	}
}

func CreateTestGraphClient() *Dygraph {
	return CreateGraphClient(testutils.GetClient(), "test")
}

func Test_marshal(t *testing.T) {
	type args struct {
		in interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]types.AttributeValue
		wantErr bool
	}{
		{
			name: "Failed to marshal",
			args: args{
				in: map[string]string{
					"": "hi",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshal(tt.args.in)
			if (err != nil) != tt.wantErr || (tt.wantErr && !errors.Is(err, MarshalError)) {
				t.Errorf("marshal() error = %v, got = %v, wantErr %v", err, got, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("marshal() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unmarshal(t *testing.T) {
	type args struct {
		m   map[string]types.AttributeValue
		out interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Failed to unmarshal",
			args: args{
				m:   nil,
				out: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := unmarshal(tt.args.m, tt.args.out); (err != nil) != tt.wantErr || tt.wantErr && !errors.Is(err, UnmarshalError) {
				t.Errorf("unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
