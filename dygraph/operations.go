package dygraph

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"log"
)

var MarshalError = errors.New("marshalling error")
var UnmarshalError = errors.New("unmarshalling error")
var DuplicateError = errors.New("duplicate")
var TooManyRequestsError = errors.New("too many requests")

type dynamoDBAPI interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Query(ctx context.Context, input *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	TransactWriteItems(ctx context.Context, input *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
}

// Dygraph type implements graph operations on top of Amazon DynamoDB
type Dygraph struct {
	client      dynamoDBAPI
	environment string
	marshal     func(in interface{}) (map[string]types.AttributeValue, error)
	unmarshal   func(m map[string]types.AttributeValue, out interface{}) error
}

func marshal(in interface{}) (map[string]types.AttributeValue, error) {
	obj, err := attributevalue.MarshalMap(in)
	if err != nil {
		log.Printf("failed to marshal a value %v into map[string]*types.AttributeValue with error %v", in, err)
		return nil, fmt.Errorf("%s: %w", err, MarshalError)
	}

	return obj, nil
}

func unmarshal(m map[string]types.AttributeValue, out interface{}) error {
	err := attributevalue.UnmarshalMap(m, out)
	if err != nil {
		log.Printf("failed to unmarshal map[string]*types.AttributeValue %v", m)
		return fmt.Errorf("%s: %w", err, UnmarshalError)
	}

	return nil
}

func wrapAwsError(err error) error {
	var conditionalCheckFailedException *types.ConditionalCheckFailedException
	if errors.As(err, &conditionalCheckFailedException) {
		log.Printf("duplicate item exception: %s", err)
		return fmt.Errorf("%s: %w", err, DuplicateError)
	}
	var tooManyRequests *types.ProvisionedThroughputExceededException
	if errors.As(err, &tooManyRequests) {
		log.Printf("too many requests exception: %s", err)
		return fmt.Errorf("%s: %w", err, DuplicateError)
	}

	return err
}

func CreateGraphClient(client dynamoDBAPI, environment string) *Dygraph {
	return &Dygraph{
		client:      client,
		environment: environment,
		marshal:     marshal,
		unmarshal:   unmarshal,
	}
}

func (r *Dygraph) getTableName() string {
	const TableName = "Authorization"

	return TableName + "-" + r.environment
}

//InsertRecord inserts a node
//TODO: Why do I pass a pointer not a value?
func (r *Dygraph) InsertRecord(node *Node) error {
	item, err := r.marshal(node.createNodeDto())

	if err != nil {
		return err
	}

	_, err = r.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		ConditionExpression: aws.String("attribute_not_exists(id)"),
		Item:                item,
		TableName:           aws.String(r.getTableName()),
	})

	if err != nil {
		return fmt.Errorf("insert record: %w", wrapAwsError(err))
	}

	return nil
}

func (r *Dygraph) GetNodes(organisationId uuid.UUID, nodeType string) ([]Node, error) {
	output, err := r.client.Query(context.TODO(), &dynamodb.QueryInput{
		IndexName:              aws.String("GSIApplicationTypeTarget"),
		TableName:              aws.String(r.getTableName()),
		KeyConditionExpression: aws.String("organisationId = :organisationId and begins_with(typeTarget, :type)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":organisationId": &types.AttributeValueMemberS{Value: organisationId.String()},
			":type":           &types.AttributeValueMemberS{Value: nodePrefix + nodeType},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("get nodes: %w", wrapAwsError(err))

	}

	result := make([]Node, output.Count)
	for i, item := range output.Items {
		d := dto{}
		err := r.unmarshal(item, &d)
		if err != nil {
			return nil, err
		}
		result[i] = d.createNode()
	}

	return result, nil
}

func (r *Dygraph) GetEdges(organisationId uuid.UUID, edgeType string) ([]Edge, error) {
	output, err := r.client.Query(context.TODO(), &dynamodb.QueryInput{
		IndexName:              aws.String("GSIApplicationTypeTarget"),
		TableName:              aws.String(r.getTableName()),
		KeyConditionExpression: aws.String("organisationId = :organisationId and begins_with(typeTarget, :type)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":organisationId": &types.AttributeValueMemberS{Value: organisationId.String()},
			":type":           &types.AttributeValueMemberS{Value: edgePrefix + edgeType},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("get edges: %w", wrapAwsError(err))
	}

	result := make([]Edge, output.Count)
	for i, item := range output.Items {
		d := dto{}
		err := r.unmarshal(item, &d)
		if err != nil {
			return nil, err
		}
		result[i] = d.createEdge()
	}

	return result, nil
}

func (r *Dygraph) GetNodeEdgesOfType(organisationId, id uuid.UUID, edgeType string) ([]Edge, error) {
	output, err := r.client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(r.getTableName()),
		KeyConditionExpression: aws.String("globalId = :globalId and begins_with(typeTarget, :type)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":globalId": &types.AttributeValueMemberS{Value: organisationId.String() + "_" + id.String()},
			":type":     &types.AttributeValueMemberS{Value: edgePrefix + edgeType},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("get node edges of type: %w", wrapAwsError(err))
	}

	result := make([]Edge, output.Count)
	for i, item := range output.Items {
		d := dto{}
		err := r.unmarshal(item, &d)
		if err != nil {
			return nil, err
		}
		result[i] = d.createEdge()
	}

	return result, nil
}

func (r *Dygraph) TransactionalInsert(items []Edge) error {
	transactWriteItems := make([]types.TransactWriteItem, len(items))
	for i := 0; i < len(items); i++ {
		av, err := r.marshal(items[i].createEdgeDto())
		if err != nil {
			return err
		}
		transactWriteItems[i] = types.TransactWriteItem{
			Put: &types.Put{
				ConditionExpression: aws.String("attribute_not_exists(id)"),
				Item:                av,
				TableName:           aws.String(r.getTableName()),
			},
		}
	}
	_, err := r.client.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: transactWriteItems,
	})

	if err != nil {
		var transactionCancelledException *types.TransactionCanceledException
		if errors.As(err, &transactionCancelledException) {
			for i, reason := range transactionCancelledException.CancellationReasons {
				if *reason.Code == "ConditionalCheckFailed" {
					log.Printf("duplicate item %v", items[i])
				}
			}
			return fmt.Errorf("duplicate item exception: %w", DuplicateError)
		}
	}

	return nil
}
