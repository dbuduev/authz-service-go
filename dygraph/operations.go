package dygraph

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	"log"
)

var MarshalError = errors.New("marshalling error")
var UnmarshalError = errors.New("unmarshalling error")

type dynamoDBAPI interface {
	PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
	TransactWriteItems(input *dynamodb.TransactWriteItemsInput) (*dynamodb.TransactWriteItemsOutput, error)
}

// Dygraph type implements graph operations on top of Amazon DynamoDB
type Dygraph struct {
	client      dynamoDBAPI
	environment string
	marshal     func(in interface{}) (map[string]*dynamodb.AttributeValue, error)
	unmarshal   func(m map[string]*dynamodb.AttributeValue, out interface{}) error
}

func marshal(in interface{}) (map[string]*dynamodb.AttributeValue, error) {
	obj, err := dynamodbattribute.MarshalMap(in)
	if err != nil {
		log.Printf("failed to marshal a value %v into map[string]*dynamodb.AttributeValue with error %v", in, err)
		return nil, fmt.Errorf("%s: %w", err, MarshalError)
	}

	return obj, nil
}

func unmarshal(m map[string]*dynamodb.AttributeValue, out interface{}) error {
	err := dynamodbattribute.UnmarshalMap(m, out)
	if err != nil {
		log.Printf("failed to unmarshal map[string]*dynamodb.AttributeValue %v", m)
		return fmt.Errorf("%s: %w", err, UnmarshalError)
	}

	return nil
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

	_, err = r.client.PutItem(&dynamodb.PutItemInput{
		ConditionExpression: aws.String("attribute_not_exists(id)"),
		Item:                item,
		TableName:           aws.String(r.getTableName()),
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *Dygraph) GetNodes(organisationId uuid.UUID, nodeType string) ([]Node, error) {
	output, err := r.client.Query(&dynamodb.QueryInput{
		IndexName:              aws.String("GSIApplicationTypeTarget"),
		TableName:              aws.String(r.getTableName()),
		KeyConditionExpression: aws.String("organisationId = :organisationId and begins_with(typeTarget, :type)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":organisationId": {
				S: aws.String(organisationId.String()),
			},
			":type": {
				S: aws.String(nodePrefix + nodeType),
			},
		},
	})

	if err != nil {
		return nil, err
	}

	result := make([]Node, *output.Count)
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
	output, err := r.client.Query(&dynamodb.QueryInput{
		IndexName:              aws.String("GSIApplicationTypeTarget"),
		TableName:              aws.String(r.getTableName()),
		KeyConditionExpression: aws.String("organisationId = :organisationId and begins_with(typeTarget, :type)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":organisationId": {
				S: aws.String(organisationId.String()),
			},
			":type": {
				S: aws.String(edgePrefix + edgeType),
			},
		},
	})

	if err != nil {
		return nil, err
	}

	result := make([]Edge, *output.Count)
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
	output, err := r.client.Query(&dynamodb.QueryInput{
		TableName:              aws.String(r.getTableName()),
		KeyConditionExpression: aws.String("globalId = :globalId and begins_with(typeTarget, :type)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":globalId": {
				S: aws.String(organisationId.String() + "_" + id.String()),
			},
			":type": {
				S: aws.String(edgePrefix + edgeType),
			},
		},
	})

	if err != nil {
		return nil, err
	}

	result := make([]Edge, *output.Count)
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
	transactWriteItems := make([]*dynamodb.TransactWriteItem, len(items))
	for i := 0; i < len(items); i++ {
		av, err := r.marshal(items[i].createEdgeDto())
		if err != nil {
			return err
		}
		transactWriteItems[i] = &dynamodb.TransactWriteItem{
			Put: &dynamodb.Put{
				ConditionExpression: aws.String("attribute_not_exists(id)"),
				Item:                av,
				TableName:           aws.String(r.getTableName()),
			},
		}
	}
	_, err := r.client.TransactWriteItems(&dynamodb.TransactWriteItemsInput{
		TransactItems: transactWriteItems,
	})

	return err
}
