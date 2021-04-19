package dygraph

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type dynamodbAPIStub struct {
	putItem            func(ctx context.Context, input *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	query              func(ctx context.Context, input *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	transactWriteItems func(ctx context.Context, input *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
}

func (d *dynamodbAPIStub) PutItem(ctx context.Context, input *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return d.putItem(ctx, input, optFns...)
}

func (d *dynamodbAPIStub) Query(ctx context.Context, input *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return d.query(ctx, input, optFns...)
}

func (d *dynamodbAPIStub) TransactWriteItems(ctx context.Context, input *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error) {
	return d.transactWriteItems(ctx, input, optFns...)
}
