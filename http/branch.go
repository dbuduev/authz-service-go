package http

import (
	"encoding/json"
	"github.com/dbuduev/authz-service-go/core"
	"github.com/dbuduev/authz-service-go/sphinx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
)

type (
	BranchRepository interface {
		AddBranch(b core.Branch) error
		AssignBranchToBranchGroup(x core.BranchAssignment) error
		GetBranchesByBranchGroup(organisationId, branchGroupId uuid.UUID) ([]uuid.UUID, error)
		GetHierarchy(organisationId uuid.UUID) (sphinx.BranchGroupContent, error)
	}
	BranchResource struct {
		repository BranchRepository
		Router     func(chi.Router)
	}
	branchCreateRequest struct {
		Id   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}
)

func (r branchCreateRequest) ToBranch(organisationId uuid.UUID) core.Branch {
	return core.Branch{
		OrganisationId: organisationId,
		Id:             r.Id,
		Name:           r.Name,
	}
}

func CreateBranchResource(repository BranchRepository) *BranchResource {
	res := &BranchResource{
		repository: repository,
		Router: func(r chi.Router) {
			r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
				writer.Write([]byte("Hi, branch"))
			})
			r.Post("/", func(writer http.ResponseWriter, request *http.Request) {
				ctx := request.Context()
				organisationId, ok := ctx.Value(OrganisationIdKey).(uuid.UUID)
				if !ok {
					http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				payload := &branchCreateRequest{}
				err := json.NewDecoder(request.Body).Decode(payload)
				if err != nil {
					http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
					return
				}
				branch := payload.ToBranch(organisationId)
				err = repository.AddBranch(branch)
				if err != nil {
					http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
					return
				}
				writer.Write([]byte("branch created"))
			})
		},
	}

	return res
}
