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

// Dygraph type implements graph operations on top of Amazon DynamoDB
type Dygraph struct {
	client      *dynamodb.DynamoDB
	environment string
	marshal     func(in interface{}) (map[string]*dynamodb.AttributeValue, error)
}

func marshal(in interface{}) (map[string]*dynamodb.AttributeValue, error) {
	obj, err := dynamodbattribute.MarshalMap(in)
	const errMessage = "failed to marshal a value"
	if err != nil {
		log.Printf("%s %v into map[string]*dynamodb.AttributeValue map", errMessage, in)
	}
	return obj, fmt.Errorf("%s: %w", errMessage, err)
}

//func unmarshal(map[string]*dynamodb.AttributeValue, in interface{}) (, error) {
//	obj, err := dynamodbattribute.MarshalMap(in)
//	const errMessage = "failed to marshal a value"
//	if err != nil {
//		log.Printf("%s %v into map[string]*dynamodb.AttributeValue map", errMessage, in)
//	}
//	return obj, fmt.Errorf("%s: %w", errMessage, err)
//}

func CreateGraphClient(client *dynamodb.DynamoDB, environment string) *Dygraph {
	return &Dygraph{
		client:      client,
		environment: environment,
		marshal:     marshal,
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
		return fmt.Errorf("failed to marshal a node: %w", err)
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
		err := dynamodbattribute.UnmarshalMap(item, &d)
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
		err := dynamodbattribute.UnmarshalMap(item, &d)
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
		err := dynamodbattribute.UnmarshalMap(item, &d)
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
