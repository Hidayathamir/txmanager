# txmanager

Gorm transaction manager

## How to use

Example [how to use](https://github.com/Hidayathamir/golang_repository_pattern_gin_gorm_sql_transaction).

Check file [service.go](https://github.com/Hidayathamir/golang_repository_pattern_gin_gorm_sql_transaction/blob/main/api/v1/payment/service/service.go). `Transfer` method do `txManager.SQLTransaction`, it can be nested like when it's called `p.updateBalanceSenderAndRecipient` which have it's own `txManager.SQLTransaction`.

```go
package service

import (
	"context"
	"errors"

	"github.com/Hidayathamir/golang_repository_pattern_gin_gorm_sql_transaction/api/v1/payment/dto"
	"github.com/Hidayathamir/golang_repository_pattern_gin_gorm_sql_transaction/api/v1/payment/repository"
	"github.com/Hidayathamir/golang_repository_pattern_gin_gorm_sql_transaction/database/model"
	"github.com/Hidayathamir/txmanager"
)

type IPaymentService interface {
	Transfer(ctx context.Context, req dto.ReqTransfer) (model.Transaction, error)
}

type PaymentService struct {
	repo      repository.IPaymentRepo
	txManager txmanager.ITransactionManager
}

func NewPaymentService(repo repository.IPaymentRepo, txManager txmanager.ITransactionManager) IPaymentService {
	return &PaymentService{repo: repo, txManager: txManager}
}

func (p *PaymentService) Transfer(ctx context.Context, req dto.ReqTransfer) (model.Transaction, error) {
	if err := validateReqTransfer(req); err != nil {
		return model.Transaction{}, err
	}

	var transaction model.Transaction
	err := p.txManager.SQLTransaction(ctx, func(ctx context.Context) error {
		sender, err := p.repo.GetUserByID(ctx, req.SenderID)
		if err != nil {
			return err
		}

		if req.Amount > sender.Balance {
			return errors.New("balance is not enough")
		}

		recipient, err := p.repo.GetUserByID(ctx, req.RecipientID)
		if err != nil {
			return err
		}

		transaction, err = p.repo.CreateTransaction(ctx, model.Transaction{
			SenderID:    req.SenderID,
			RecipientID: req.RecipientID,
			Amount:      req.Amount,
		})
		if err != nil {
			return err
		}

		err = p.updateBalanceSenderAndRecipient(ctx, req.Amount, sender, recipient)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return model.Transaction{}, err
	}

	return transaction, nil
}

func validateReqTransfer(req dto.ReqTransfer) error {
	if req.SenderID == 0 {
		return errors.New("sender_id can not be empty")
	}

	if req.RecipientID == 0 {
		return errors.New("recipient_id can not be empty")
	}

	if req.SenderID == req.RecipientID {
		return errors.New("can not transfer to yourself")
	}

	if req.Amount <= 10000 {
		return errors.New("amount can not be less than 10000")
	}

	return nil
}

func (p *PaymentService) updateBalanceSenderAndRecipient(ctx context.Context, transferAmount int, sender model.User, recipient model.User) error {
	// txManager.SQLTransaction can be nested
	return p.txManager.SQLTransaction(ctx, func(ctx context.Context) error {
		err := p.repo.UpdateUserBalance(ctx, sender.ID, sender.Balance-transferAmount)
		if err != nil {
			return err
		}

		err = p.repo.UpdateUserBalance(ctx, recipient.ID, recipient.Balance+transferAmount)
		if err != nil {
			return err
		}

		return nil
	})
}
```