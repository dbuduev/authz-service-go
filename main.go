package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/dbuduev/authz-service-go/dygraph"
	resource "github.com/dbuduev/authz-service-go/http"
	"github.com/dbuduev/authz-service-go/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"net/http"
	"time"
)

func GetClient() *dynamodb.DynamoDB {
	// Create DynamoDB client
	s := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return dynamodb.New(s, s.Config.WithEndpoint("http://localhost:8000"))
}

func main() {
	repo := repository.CreateRepository(dygraph.CreateGraphClient(GetClient(), "test"))
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route(fmt.Sprintf("/{%s}", resource.OrganisationIdKey), func(r chi.Router) {
		r.Use(organisationContext)
		r.Route("/branch", resource.CreateBranchResourceRouter(repo))
		r.Get("/branch-group", func(writer http.ResponseWriter, request *http.Request) {
			ctx := request.Context()
			organisationId, ok := ctx.Value(resource.OrganisationIdKey).(uuid.UUID)
			if !ok {
				http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			writer.Write([]byte(fmt.Sprintf("organisation:%s", organisationId)))
		})
	})
	server := http.Server{
		Addr:         ":8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      r,
	}
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func organisationContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		organisationId, err := uuid.Parse(chi.URLParam(r, resource.OrganisationIdKey))
		if err != nil {
			http.Error(w, "Can't parse organisation id, must be UUID.", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), resource.OrganisationIdKey, organisationId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
