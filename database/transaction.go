package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var ErrConditionalCheckFailed = fmt.Errorf("ConditionalCheckFailed")

type Transaction struct {
	transactionResults []*transactionResult
}

type transactionResult struct {
	transactionWriteItem types.TransactWriteItem
	err                  error
}

func NewTransaction() *Transaction {
	return &Transaction{
		transactionResults: []*transactionResult{},
	}
}

func (transaction *Transaction) Include(writeItem types.TransactWriteItem, err error) (tr *Transaction) {
	result := &transactionResult{
		transactionWriteItem: writeItem,
		err:                  err,
	}
	transaction.transactionResults = append(transaction.transactionResults, result)
	return transaction
}

func (transaction *Transaction) Execute(ctx context.Context, dynamodbClient *dynamodb.Client) (err error) {
	transactionWriteItems := []types.TransactWriteItem{}
	for _, result := range transaction.transactionResults {
		if result.err != nil {
			err = result.err
			return
		}
		transactionWriteItems = append(transactionWriteItems, result.transactionWriteItem)
	}

	transactionWriteItemsInput := &dynamodb.TransactWriteItemsInput{
		TransactItems: transactionWriteItems,
	}
	_, err = dynamodbClient.TransactWriteItems(ctx, transactionWriteItemsInput)

	var conditionalCheckFailedException *types.ConditionalCheckFailedException
	if errors.As(err, &conditionalCheckFailedException) {
		err = ErrConditionalCheckFailed
		return
	}

	var transactionCanceledException *types.TransactionCanceledException
	if errors.As(err, &transactionCanceledException) {
		for _, reason := range transactionCanceledException.CancellationReasons {
			if reason.Code != nil && *reason.Code == "ConditionalCheckFailed" {
				err = ErrConditionalCheckFailed
				return
			}
		}
	}

	return
}
