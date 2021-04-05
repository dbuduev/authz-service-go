// +build integration

package http

import (
	"net/http/httptest"
	"testing"
)

func TestBranchesAndBranchGroups(t *testing.T) {
	httptest.NewServer(ConfigureH)
}
