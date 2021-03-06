package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/dbuduev/authz-service-go/dygraph"
	resource "github.com/dbuduev/authz-service-go/http"
	"github.com/dbuduev/authz-service-go/repository"
	"log"
	"net/http"
	"time"
)

func GetClient() *dynamodb.Client {
	// Create DynamoDB client
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigFiles(config.DefaultSharedConfigFiles),
		config.WithSharedCredentialsFiles(config.DefaultSharedCredentialsFiles),
	)
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		//o.ClientLogMode |= aws.LogRequestWithBody
		o.EndpointResolver = dynamodb.EndpointResolverFunc(
			func(region string, options dynamodb.EndpointResolverOptions) (aws.Endpoint, error) {
				options.DisableHTTPS = true
				return aws.Endpoint{URL: "http://localhost:8000", HostnameImmutable: true}, nil
			})
	})
}

func main() {
	repo := repository.CreateRepository(dygraph.CreateGraphClient(GetClient(), "test"))

	server := http.Server{
		Addr:         ":8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      resource.ConfigureHandler(repo),
	}
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
