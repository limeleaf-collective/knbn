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

var getAllHashes = redis.NewScript(`
local key = KEYS[1]
local hashKeys = {}
local hashes = {}
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
		hashKeys[#hashKeys+1] = v
	end
until tonumber(cursor) == 0

local hashesIdx = 1
for i, _ in ipairs(hashKeys) do
	local hashKey = hashKeys[i]

	local hash = redis.pcall("HGETALL", hashKey)
	if type(hash) == 'table' and hash.err then
		err = "hgetall failed '" .. hashKey .. "'"
		redis.log(redis.LOG_NOTICE, err, result.err)
	end

	if hash[1] then
		hashes[hashesIdx] = hashKey
		hashes[hashesIdx + 1] = cjson.encode(hash)
		hashesIdx = hashesIdx + 2
	end
end

if err ~= nil then
	return { err = err }
end

return hashes
`)

func (r Redis) getHashesForKey(ctx context.Context, key string) ([]map[string]string, error) {
	var hashes []map[string]string
	raw, err := getAllHashes.Run(ctx, r.db, []string{key}).StringSlice()
	if err != nil {
		return hashes, errs.Join(ErrStorage, err)
	}

	for i := 0; i < len(raw); i += 2 {
		listKey := raw[i]
		listJSON := raw[i+1]

		var props []string
		if err := json.Unmarshal([]byte(listJSON), &props); err != nil {
			return hashes, errs.Join(ErrStorage, err)
		}

		id := strings.Split(listKey, ":")[1]

		hash := make(map[string]string)
		hash["id"] = id
		for pi := 0; pi < len(props); pi += 2 {
			k := props[pi]
			v := props[pi+1]
			hash[k] = v
		}

		hashes = append(hashes, hash)
	}

	return hashes, nil
}

func (r Redis) GetBoards(ctx context.Context) ([]data.Board, error) {
	var boards []data.Board
	bms, err := r.getHashesForKey(ctx, boardKey("*"))
	if err != nil {
		return boards, errs.Join(ErrStorage, err)
	}

	for _, b := range bms {
		board := data.Board{}
		for k, v := range b {
			switch k {
			case "id":
				board.ID = v
			case "title":
				board.Title = v
			case "lists":
				ids := strings.Split(v, ",")
				if ids[0] != "" {
					board.ListIDs = ids
				}
			default:
				slog.Warn("unknown board property", "prop", k, "value", v)
			}
		}

		boards = append(boards, board)
	}

	return boards, nil
}

func (r Redis) GetListsForBoard(ctx context.Context, id string) ([]data.List, error) {
	var lists []data.List
	lms, err := r.getHashesForKey(ctx, listKey(id, "*"))
	if err != nil {
		return lists, errs.Join(ErrStorage, err)
	}

	for _, b := range lms {
		list := data.List{}
		for k, v := range b {
			switch k {
			case "id":
				ids := strings.Split(v, ":")
				list.ID = ids[len(ids)-1]
			case "title":
				list.Title = v
			case "cards":
				ids := strings.Split(v, ",")
				if ids[0] != "" {
					list.CardIDs = ids
				}
			default:
				slog.Warn("unknown list property", "prop", k, "value", v)
			}
		}

		lists = append(lists, list)
	}

	return lists, nil
}

func (r Redis) SetListsForBoard(ctx context.Context, boardID string, listIDs []string) error {
	if boardID == "" {
		return errors.Wrap(ErrInvalid, "missing id")
	}
	k := boardKey(boardID)

	exists, err := r.db.Exists(ctx, k).Result()
	if err != nil {
		return errs.Join(ErrStorage, err)
	} else if exists == 0 {
		return ErrNotFound
	}

	lists := strings.Join(listIDs, ",")
	if err := r.db.HSet(ctx, k, "lists", lists).Err(); err != nil {
		return errs.Join(ErrStorage, err)
	}

	return nil
}

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

func listKey(bid, lid string) string {
	return fmt.Sprintf("lists:%s:%s", bid, lid)
}

//func cardKey(lid, cid string) string {
//	return fmt.Sprintf("cards:%s:%s", lid, cid)
//
//}
