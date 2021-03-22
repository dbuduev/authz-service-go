package repository

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

type Repository struct {
	client      *dynamodb.DynamoDB
	environment string
}

func CreateRepository(client *dynamodb.DynamoDB, environment string) *Repository {
	return &Repository{
		client:      client,
		environment: environment,
	}
}

func (r *Repository) getTableName() string {
	const TableName = "Authorization"

	return TableName + "-" + r.environment
}

func (r *Repository) insertNode(node *LogicalRecordRequest) error {
	dto := node.createNodeDto()
	av, err := dynamodbattribute.MarshalMap(dto)

	if err != nil {
		return err
	}

	_, err = r.client.PutItem(&dynamodb.PutItemInput{
		ConditionExpression: aws.String("attribute_not_exists(id)"),
		Item:                av,
		TableName:           aws.String(r.getTableName()),
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) getNodes(organisationId uuid.UUID, nodeType string) ([]LogicalRecord, error) {
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

	result := make([]LogicalRecord, *output.Count)
	for i, item := range output.Items {
		var nodeDto = dto{}
		err := dynamodbattribute.UnmarshalMap(item, &nodeDto)
		if err != nil {
			return nil, err
		}
		result[i] = nodeDto.createLogicalRecord()
	}

	return result, nil
}

func (r *Repository) getEdges(organisationId uuid.UUID, edgeType string) ([]LogicalRecord, error) {
	output, err := r.client.Query(&dynamodb.QueryInput{
		IndexName:              aws.String("GSIApplicationTypeTarget"),
		TableName:              aws.String(r.getTableName()),
		KeyConditionExpression: aws.String("organisationId = :organisationId and begins_with(typeTarget, :type)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":organisationId": {
				S: aws.String(organisationId.String()),
			},
			":type": {
				S: aws.String(nodePrefix + edgeType),
			},
		},
	})

	if err != nil {
		return nil, err
	}

	result := make([]LogicalRecord, *output.Count)
	for i, item := range output.Items {
		var dto = dto{}
		err := dynamodbattribute.UnmarshalMap(item, &dto)
		if err != nil {
			return nil, err
		}
		result[i] = dto.createLogicalRecord()
	}

	return result, nil
}

func (r *Repository) getNodeEdgesOfType(organisationId, id uuid.UUID, edgeType string) ([]LogicalRecord, error) {
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

	result := make([]LogicalRecord, *output.Count)
	for i, item := range output.Items {
		var dto = dto{}
		err := dynamodbattribute.UnmarshalMap(item, &dto)
		if err != nil {
			return nil, err
		}
		result[i] = dto.createLogicalRecord()
	}

	return result, nil
}

func (r *Repository) transactionalInsert(items []CreateEdgeRequest) error {
	transactWriteItems := make([]*dynamodb.TransactWriteItem, len(items))
	for i := 0; i < len(items); i++ {
		av, err := dynamodbattribute.MarshalMap(items[i].createNodeDto())
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
