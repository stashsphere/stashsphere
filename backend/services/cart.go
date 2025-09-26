package services

import (
	"context"
	"database/sql"

	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/utils"
)

type CartService struct {
	db *sql.DB
}

func NewCartService(db *sql.DB) *CartService {
	return &CartService{db}
}

type UpdateCartParams struct {
	UserId   string
	ThingIds []string
}

func (cs *CartService) UpdateCart(ctx context.Context, params UpdateCartParams) (models.CartEntrySlice, error) {
	var outerCartEntries models.CartEntrySlice
	err := utils.Tx(ctx, cs.db, func(tx *sql.Tx) error {
		err := operations.UpdateCart(ctx, tx, params.ThingIds, params.UserId)
		if err != nil {
			return err
		}
		entries, err := operations.GetCart(ctx, tx, params.UserId)
		if err != nil {
			return err
		}
		outerCartEntries = entries
		return nil
	})
	if err != nil {
		return nil, err
	}
	return outerCartEntries, nil
}

func (cs *CartService) GetCart(ctx context.Context, userId string) (models.CartEntrySlice, error) {
	return operations.GetCart(ctx, cs.db, userId)
}
