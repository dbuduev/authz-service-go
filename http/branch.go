package http

import (
	"encoding/json"
	"github.com/dbuduev/authz-service-go/core"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
)

type (
	BranchRepository interface {
		AddBranch(b core.Branch) error
	}
	BranchResource struct {
		repository BranchRepository
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

func (r BranchResource) AddBranch() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
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
		err = r.repository.AddBranch(branch)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		writer.Write([]byte("branch created"))
	}
}

func CreateBranchResourceRouter(repository BranchRepository) func(r chi.Router) {
	res := &BranchResource{repository: repository}

	return func(r chi.Router) {
		r.Post("/", res.AddBranch())
	}
}
