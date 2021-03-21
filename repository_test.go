package repository

import (
	"fmt"
	"github.com/google/uuid"
	"testing"
)

func TestGetNodes(t *testing.T) {
	node := Node{
		OrganisationId: uuid.New(),
		Id:             uuid.New(),
		Type:           "ROLE",
		Data:           "Branch manager",
	}
	err := InsertNode(&node)
	if err != nil {
		t.Fatalf("Failed to insert node %v with error %v", node, err)
	}

	result, err := GetNodes(node.OrganisationId, node.Type)
	if err != nil {
		t.Fatalf("Failed to get nodes with error %v", err)
	}

	if node != result[0] {
		t.Errorf("Expect %s, got %s", node, result[0])
	}
}

func ExampleString() {
	s := "world"
	f := func(a string) bool {
		return &a == &s
	}
	fmt.Println(f(s))
	// Output: true1
}
