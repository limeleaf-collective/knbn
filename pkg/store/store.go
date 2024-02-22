package store

import (
	"context"
	"github.com/limeleaf-coop/knbn/pkg/data"
)

type Store interface {
	CreateBoard(ctx context.Context, board data.Board) error
	GetBoard(ctx context.Context, id string) (data.Board, error)
	UpdateBoard(ctx context.Context, board data.Board) error
	DeleteBoard(ctx context.Context, id string) error
	GetBoards(ctx context.Context) ([]data.Board, error)
	GetListsForBoard(ctx context.Context, id string) ([]data.List, error)
	SetListsForBoard(ctx context.Context, boardID string, listIDs []string) error

	CreateList(ctx context.Context, list data.List) error
	GetList(ctx context.Context, id string) (data.List, error)
	UpdateList(ctx context.Context, list data.List) error
	DeleteList(ctx context.Context, id string) error
	GetCardsForList(ctx context.Context, id string) ([]data.Card, error)
	SetCardsForList(ctx context.Context, listID string, cardIDs []string) error

	CreateCard(ctx context.Context, id data.Card) error
	UpdateCard(ctx context.Context, id data.Card) error
	DeleteCard(ctx context.Context, slug string) error
}
