package http

import (
	"context"
	"fmt"
	"github.com/dbuduev/authz-service-go/core"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"net/http"
)

const (
	OrganisationIdKey = "organisationId"
	BranchGroupIdKey  = "branchGroupId"
)

func ConfigureHandler(repo core.Repository) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route(fmt.Sprintf("/{%s}", OrganisationIdKey), func(r chi.Router) {
		r.Use(organisationContext)
		r.Route("/branch", CreateBranchResourceRouter(repo))
		r.Route("/branch-group", CreateBranchGroupResourceRouter(repo))
	})
	return r
}

func organisationContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		organisationId, err := uuid.Parse(chi.URLParam(r, OrganisationIdKey))
		if err != nil {
			http.Error(w, "Can't parse organisation id, must be UUID.", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), OrganisationIdKey, organisationId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
