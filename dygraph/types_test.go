package dygraph

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"testing"
)

func Test_dto_createEdge(t *testing.T) {
	tests := []struct {
		name string
		want Edge
	}{
		{"Tagless case", Edge{
			OrganisationId: uuid.New(),
			Id:             uuid.New(),
			TargetNodeId:   uuid.New(),
			TargetNodeType: "ROLE",
			Tags:           nil,
			Data:           uuid.New().String(),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.want.createEdgeDto()
			got := d.createEdge()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("createEdge() diff %v", diff)
			}
		})
	}
}

func Test_dto_createNode(t *testing.T) {
	tests := []struct {
		name string
		want Node
	}{
		{"node creation test", Node{
			OrganisationId: uuid.New(),
			Id:             uuid.New(),
			Type:           "OP",
			Data:           "manage-staff",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.want.createNodeDto()
			got := d.createNode()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("createNode() diff %v", diff)
			}
		})
	}
}
