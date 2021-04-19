package testutils

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log"
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
