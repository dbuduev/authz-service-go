package dygraph

import "github.com/aws/aws-sdk-go/service/dynamodb"

type dynamodbAPIStub struct {
	putItem            func(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	query              func(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
	transactWriteItems func(input *dynamodb.TransactWriteItemsInput) (*dynamodb.TransactWriteItemsOutput, error)
}

func (d *dynamodbAPIStub) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return d.putItem(input)
}

func (d *dynamodbAPIStub) Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	return d.query(input)
}

func (d *dynamodbAPIStub) TransactWriteItems(input *dynamodb.TransactWriteItemsInput) (*dynamodb.TransactWriteItemsOutput, error) {
	return d.transactWriteItems(input)
}
