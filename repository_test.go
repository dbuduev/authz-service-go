package repository

import "testing"

func TestGetNodes(t *testing.T) {
	node := Node{
		OrganisationId: "organisationId",
		Id:             "id",
		Type:           "ROLE",
		Data:           "Branch manager",
	}
	CreateNode(&node)
	result := GetNodes(node.OrganisationId, node.Type)
	if node != result[0] {
		t.Errorf("Expect %s, got %s", node, result[0])
	}
}
