package http

import (
	"encoding/json"
	"fmt"
	"github.com/dbuduev/authz-service-go/core"
	"github.com/dbuduev/authz-service-go/sphinx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"io"
	"net/http"
)

type (
	BranchGroupRepository interface {
		AddBranchGroup(g core.BranchGroup) error
		AssignBranchToBranchGroup(x core.BranchAssignment) error
		GetBranchesByBranchGroup(organisationId, branchGroupId uuid.UUID) ([]uuid.UUID, error)
		GetHierarchy(organisationId uuid.UUID) (sphinx.BranchGroupContent, error)
	}
	BranchGroupResource struct {
		repository BranchGroupRepository
	}
	branchGroupCreateRequest struct {
		Id   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}
	assignBranchRequest struct {
		BranchId uuid.UUID `json:"branch_id"`
	}
)

func (r branchGroupCreateRequest) To(organisationId uuid.UUID) core.BranchGroup {
	return core.BranchGroup{
		OrganisationId: organisationId,
		Id:             r.Id,
		Name:           r.Name,
	}
}

func (r assignBranchRequest) To(organisationId, branchGroupId uuid.UUID) core.BranchAssignment {
	return core.BranchAssignment{
		OrganisationId: organisationId,
		BranchId:       r.BranchId,
		BranchGroupId:  branchGroupId,
	}
}

func (r BranchGroupResource) AddBranchGroup() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		organisationId, ok := ctx.Value(OrganisationIdKey).(uuid.UUID)
		if !ok {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		payload := &branchGroupCreateRequest{}
		err := json.NewDecoder(request.Body).Decode(payload)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		branchGroup := payload.To(organisationId)
		err = r.repository.AddBranchGroup(branchGroup)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		_, _ = writer.Write([]byte("branchGroup created"))
	}
}

func (r BranchGroupResource) AssignBranchToBranchGroup() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		organisationId, ok := ctx.Value(OrganisationIdKey).(uuid.UUID)
		if !ok {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		branchGroupId, err := uuid.Parse(chi.URLParam(request, BranchGroupIdKey))
		if err != nil {
			http.Error(writer, fmt.Sprintf("%s should UUID", BranchGroupIdKey), http.StatusBadRequest)
			return
		}
		payload := &assignBranchRequest{}
		err = json.NewDecoder(request.Body).Decode(payload)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		branchAssignment := payload.To(organisationId, branchGroupId)
		err = r.repository.AssignBranchToBranchGroup(branchAssignment)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		_, _ = io.WriteString(writer, "branchAssignment created")
	}
}

func (r BranchGroupResource) GetBranchesByBranchGroup() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		organisationId, ok := ctx.Value(OrganisationIdKey).(uuid.UUID)
		if !ok {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		branchGroupId, err := uuid.Parse(chi.URLParam(request, BranchGroupIdKey))
		if err != nil {
			http.Error(writer, fmt.Sprintf("%s should UUID", BranchGroupIdKey), http.StatusBadRequest)
			return
		}
		branches, err := r.repository.GetBranchesByBranchGroup(organisationId, branchGroupId)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(writer).Encode(branches)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}
}

func CreateBranchGroupResourceRouter(repository BranchGroupRepository) func(r chi.Router) {
	res := &BranchGroupResource{repository: repository}

	return func(r chi.Router) {
		r.Post("/", res.AddBranchGroup())
		r.Put(fmt.Sprintf("/{%s}", BranchGroupIdKey), res.AssignBranchToBranchGroup())
		r.Get(fmt.Sprintf("/{%s}", BranchGroupIdKey), res.GetBranchesByBranchGroup())
	}
}
