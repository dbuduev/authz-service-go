package dygraph

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
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

func TestDygraph_InsertRecord(t *testing.T) {
	graphClient := CreateTestGraphClient()
	node := Node{
		OrganisationId: uuid.New(),
		Id:             uuid.New(),
		Type:           "ROLE",
		Data:           "Branch manager",
	}
	err := graphClient.InsertRecord(&node)
	if err != nil {
		t.Fatalf("Failed to insert a record with error: %v", err)
	}
	err = graphClient.InsertRecord(&node)
	fmt.Println("Error" + err.Error())
	if err == nil {
		t.Errorf("expected an error upon inserting a duplicate")
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

func GetClient() *dynamodb.DynamoDB {
	// Create DynamoDB client
	s := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return dynamodb.New(s, s.Config.WithEndpoint("http://localhost:8000"))
}

func CreateTestGraphClient() *Dygraph {
	return CreateGraphClient(GetClient(), "test")
}
