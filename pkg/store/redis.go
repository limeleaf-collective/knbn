package store

import (
	"context"
	"encoding/json"
	errs "errors"
	"fmt"
	"github.com/limeleaf-coop/knbn/pkg/data"
	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"log"
	"log/slog"
	"strings"
)

type Redis struct {
	db *redis.Client
}

func NewRedis(db *redis.Client) *Redis {
	return &Redis{
		db: db,
	}
}

func (r Redis) CreateBoard(ctx context.Context, board data.Board) error {
	if board.Title == "" {
		return errors.Wrap(ErrInvalid, "missing title")
	}

	var err error
	board.ID, err = nanoid.New()
	if err != nil {
		return errs.Join(ErrStorage, err)
	}

	exists, err := r.db.Exists(ctx, boardKey(board.ID)).Result()
	if err != nil {
		return errs.Join(ErrStorage, err)
	} else if exists > 0 {
		return errors.Wrap(ErrDuplicate, "board already exists")
	}

	if err := r.db.HSet(ctx, boardKey(board.ID), "title", board.Title).Err(); err != nil {
		return errs.Join(ErrStorage, err)
	}

	return nil
}

func (r Redis) GetBoard(ctx context.Context, id string) (data.Board, error) {
	b := data.Board{
		ID: id,
	}

	if id == "" {
		return b, errors.Wrap(ErrInvalid, "missing id")
	}

	res, err := r.db.HGetAll(ctx, boardKey(id)).Result()
	if errors.Is(err, redis.Nil) {
		return b, ErrNotFound
	} else if err != nil {
		return b, errs.Join(ErrStorage, err)
	}

	b.Title = res["title"]
	b.ListIDs = strings.Split(res["lists"], ",")

	r.db.Pipeline()

	return b, nil
}

func (r Redis) UpdateBoard(ctx context.Context, board data.Board) error {
	if board.Title == "" {
		return errors.Wrap(ErrInvalid, "missing title")
	}

	if board.ID == "" {
		return errors.Wrap(ErrInvalid, "missing id")
	}

	if err := r.db.HMSet(ctx, boardKey(board.ID), "title", board.Title, "lists", strings.Join(board.ListIDs, ",")).Err(); err != nil {
		return errs.Join(ErrStorage, err)
	}

	return nil
}

func (r Redis) DeleteBoard(ctx context.Context, id string) error {
	if id == "" {
		return errors.Wrap(ErrInvalid, "missing id")
	}

	if err := r.db.HDel(ctx, boardKey(id)).Err(); err != nil {
		return errs.Join(ErrStorage, err)
	}

	return nil
}

var getBoardsScript = redis.NewScript(`
local key = KEYS[1]
local boardKeys = {}
local boards = {}
local err = nil

local cursor = 0
repeat
	local r = redis.pcall("SCAN", cursor, "MATCH", key, "COUNT", 10)
	local data
	cursor, data = unpack(r)
	if type(data) == 'table' and data.err then
		err = "scan failed '" .. key .. "'"
		redis.log(redis.LOG_NOTICE, err, result.err)
	end

	for _, v in ipairs(data) do
		boardKeys[#boardKeys+1] = v
	end
until tonumber(cursor) == 0

local boardsIdx = 1
for i, _ in ipairs(boardKeys) do
	local boardKey = boardKeys[i]

	local board = redis.pcall("HGETALL", boardKey)
	if type(board) == 'table' and board.err then
		err = "hgetall failed '" .. boardKey .. "'"
		redis.log(redis.LOG_NOTICE, err, result.err)
	end

	if board[1] then
		boards[boardsIdx] = boardKey
		boards[boardsIdx + 1] = cjson.encode(board)
		boardsIdx = boardsIdx + 2
	end
end

if err ~= nil then
	return { err = err }
end

return boards
`)

func (r Redis) GetBoards(ctx context.Context) ([]data.Board, error) {
	var boards []data.Board
	raw, err := getBoardsScript.Run(ctx, r.db, []string{boardKey("*")}).StringSlice()
	if err != nil {
		return boards, errs.Join(ErrStorage, err)
	}

	for i := 0; i < len(raw); i += 2 {
		boardKey := raw[i]
		boardJSON := raw[i+1]

		var props []string
		if err := json.Unmarshal([]byte(boardJSON), &props); err != nil {
			log.Fatalf("%+v", err)
		}

		id := strings.Split(boardKey, ":")[1]

		board := data.Board{ID: id}
		for pi := 0; pi < len(props); pi += 2 {
			switch props[pi] {
			case "title":
				board.Title = props[pi+1]
			case "lists":
				board.ListIDs = strings.Split(props[pi+1], ",")
			default:
				slog.Warn("unknown board property", "prop", props[pi], "value", props[pi+1])
			}
		}

		boards = append(boards, board)
	}

	return boards, nil
}

//func (r Redis) SetListsForBoard(ctx context.Context, boardID string, listIDs []string) error {
//
//}
//
//func (r Redis) GetListsForBoard(ctx context.Context, id string) ([]data.List, error) {
//
//}
//
//func (r Redis) CreateList(ctx context.Context, list data.List) error {
//
//}
//
//func (r Redis) GetList(ctx context.Context, id string) (data.List, error) {
//
//}
//
//func (r Redis) UpdateList(ctx context.Context, list data.List) error {
//
//}
//
//func (r Redis) DeleteList(ctx context.Context, id string) error {
//
//}
//
//func (r Redis) SetCardsForList(ctx context.Context, listID string, cardIDs []string) error {
//
//}
//
//func (r Redis) GetCardsForList(ctx context.Context, id string) ([]data.Card, error) {
//
//}
//
//func (r Redis) CreateCard(ctx context.Context, card data.Card) error {
//
//}
//
//func (r Redis) UpdateCard(ctx context.Context, card data.Card) error {
//
//}
//
//func (r Redis) DeleteCard(ctx context.Context, id string) error {
//
//}

func boardKey(id string) string {
	return fmt.Sprintf("boards:%s", id)
}

//func listSortKey(bid string) string {
//	return fmt.Sprintf("zlists:%s", bid)
//}

//func listKey(bid, lid string) string {
//	return fmt.Sprintf("lists:%s:%s", bid, lid)
//}
//
//func cardKey(lid, cid string) string {
//	return fmt.Sprintf("cards:%s:%s", lid, cid)
//
//}
//
//func cardSortKey(lid string) string {
//	return fmt.Sprintf("zcards:%s", lid)
//}
