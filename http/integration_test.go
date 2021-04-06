package http

import (
	"bytes"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/dbuduev/authz-service-go/dygraph"
	"github.com/dbuduev/authz-service-go/repository"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
)

type testClient struct {
	client *http.Client
	url    string
	t      *testing.T
}

func (c *testClient) AddBranch(branch branchCreateRequest) {
	buf, _ := json.Marshal(branch)
	response, err := c.client.Post(c.url+"/branch", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		c.t.Fatal(err)
	}
	defer response.Body.Close()
}

func (c *testClient) AddBranchGroup(branchGroup branchGroupCreateRequest) {
	buf, _ := json.Marshal(branchGroup)
	res, err := c.client.Post(c.url+"/branch-group", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		c.t.Fatal(err)
	}
	res.Body.Close()
}

func (c *testClient) AssignBranchToBranchGroup(branchGroupId uuid.UUID, branch assignBranchRequest) {
	buf, _ := json.Marshal(branch)
	req, err := http.NewRequest(http.MethodPut, c.url+"/branch-group/"+branchGroupId.String(), bytes.NewBuffer(buf))
	res, err := c.client.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	res.Body.Close()
}

func (c *testClient) GetBranchesByBranchGroup(branchGroupId uuid.UUID) []uuid.UUID {
	res, err := c.client.Get(c.url + "/branch-group/" + branchGroupId.String())
	if err != nil {
		c.t.Fatal(err)
	}
	defer res.Body.Close()
	var result []uuid.UUID
	dec := json.NewDecoder(res.Body)
	// TODO: Is this the right way to read an array?
	// read [
	_, err = dec.Token()
	if err != nil {
		c.t.Fatal(err)
	}

	// read array
	for dec.More() {
		var id uuid.UUID
		err := dec.Decode(&id)
		if err != nil {
			c.t.Fatal(err)
		}
		result = append(result, id)
	}
	// read ]
	_, err = dec.Token()
	if err != nil {
		c.t.Fatal(err)
	}

	return result
}

func TestBranchesAndBranchGroups(t *testing.T) {
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

	server := httptest.NewServer(ConfigureHandler(CreateTestRepository()))
	orgId := uuid.New()
	client := &testClient{server.Client(), server.URL + "/" + orgId.String(), t}

	albany := branchCreateRequest{uuid.New(), "Albany"}
	client.AddBranch(albany)
	milford := branchCreateRequest{uuid.New(), "Milford"}
	client.AddBranch(milford)

	branchGroup := branchGroupCreateRequest{uuid.New(), "Auckland"}
	client.AddBranchGroup(branchGroup)

	client.AssignBranchToBranchGroup(branchGroup.Id, assignBranchRequest{BranchId: albany.Id})
	client.AssignBranchToBranchGroup(branchGroup.Id, assignBranchRequest{BranchId: milford.Id})

	result := client.GetBranchesByBranchGroup(branchGroup.Id)
	want := []uuid.UUID{albany.Id, milford.Id}
	if diff := cmp.Diff(want, result, trans); diff != "" {
		t.Errorf("Received branches vs expected branches %v", diff)
	} else {
		t.Logf("Success: received branches %v. Want %v", result, want)
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

func CreateTestRepository() *repository.Repository {
	return repository.CreateRepository(CreateTestGraphClient())
}
