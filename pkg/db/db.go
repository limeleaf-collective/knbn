// Package litedoc provides a document database that uses SQLite as it's storage
// engine.
package db

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	_ "modernc.org/sqlite"
)

const (
	sqlCreateTable = "CREATE TABLE IF NOT EXISTS %s (id TEXT PRIMARY KEY, data JSON)"
	sqlInsert      = "INSERT INTO %s (id, data) VALUES ('%s', '%s')"
	sqlUpdate      = "UPDATE %s SET data = '%s' WHERE (id = '%s')"
	sqlSelect      = "SELECT data FROM %s WHERE (id = '%s')"
	sqlSelectAll   = "SELECT id, data FROM %s ORDER BY id"
	sqlQuery       = "SELECT %s.id, %s.data FROM %s, json_tree(%s.data) WHERE (fullkey LIKE '%s' AND value %s %v)"
	sqlDelete      = "DELETE FROM %s WHERE (id = '%s')"

	// Pulled from PocketBase.io for how it opens a SQLite connection.
	//
	// Note: the busy_timeout pragma must be first because
	// the connection needs to be set to block on busy before WAL mode
	// is set in case it hasn't been already set by another connection.
	pragmas = "?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=journal_size_limit(200000000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=temp_store(MEMORY)&_pragma=cache_size(-16000)"
)

// Op is a comparison operator when querying for documents.
type Op int

const (
	OpEqual Op = iota
	OpNotEqual
	OpLessThan
	OpLessThanEqual
	OpGreaterThan
	OpGreaterThanEqual
)

func (op Op) String() string {
	switch op {
	case OpEqual:
		return "="
	case OpNotEqual:
		return "!="
	case OpLessThan:
		return "<"
	case OpLessThanEqual:
		return "<="
	case OpGreaterThan:
		return ">"
	case OpGreaterThanEqual:
		return ">="
	default:
		return ""
	}
}

// Database holds the underlying SQLite database connection.
type Database struct {
	sqlite *sql.DB
}

// Open create a SQLite connection at the specified path location.
func Open(path string) (*Database, error) {
	db, err := sql.Open("sqlite", path+pragmas)
	if err != nil {
		return nil, err
	}

	return &Database{
		sqlite: db,
	}, nil
}

// Close calls Close on the underlying database.
func (db *Database) Close() error {
	return db.sqlite.Close()
}

func (db *Database) SeedFromDir(ctx context.Context, dir string) error {
	collections, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, collection := range collections {
		col := db.Collection(collection.Name())

		_, err := col.database.sqlite.ExecContext(ctx, fmt.Sprintf(sqlCreateTable, col.ID))
		if err != nil {
			return err
		}

		documents, err := os.ReadDir(path.Join(dir, collection.Name()))
		if err != nil {
			return err
		}

		for _, document := range documents {
			if document.IsDir() {
				continue
			}

			doc := col.Document(strings.TrimSuffix(document.Name(), ".json"))

			docPath := path.Join(dir, collection.Name(), document.Name())
			data, err := os.ReadFile(docPath)
			if err != nil {
				return err
			}

			doc.data = data

			_, err = doc.collection.database.sqlite.ExecContext(ctx, fmt.Sprintf(sqlInsert, doc.collection.ID, doc.ID, doc.data))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Collection returns a reference to a database collection.
func (db *Database) Collection(id string) *Collection {
	return &Collection{
		database: db,
		ID:       id,
	}
}

// Collection represents the top level structure that holds Documents.
type Collection struct {
	database *Database
	ID       string
}

// Document returns a reference to a Document within the Collection.
func (c *Collection) Document(id string) *Document {
	return &Document{
		collection: c,
		ID:         id,
	}
}

func (c *Collection) QueryAll(ctx context.Context) ([]*Document, error) {
	sql := fmt.Sprintf(sqlSelectAll, c.ID)

	r, err := c.database.sqlite.QueryContext(ctx, sql)
	if err != nil {
		return nil, err
	}

	docs := make([]*Document, 0)
	for r.Next() {
		if err := r.Err(); err != nil {
			return nil, err
		}

		doc := Document{}

		if err := r.Scan(&doc.ID, &doc.data); err != nil {
			return nil, err
		}

		docs = append(docs, &doc)
	}

	return docs, nil
}

// Query returns a list of Documents where the values at keypath match the value
// based on the Op used.
func (c *Collection) Query(ctx context.Context, keypath string, op Op, val any) ([]*Document, error) {
	switch v := val.(type) {
	case string:
		val = "'" + v + "'"
	case []byte:
		val = "'" + string(v) + "'"
	case bool:
		val = 0
		if v {
			val = 1
		}
	}

	sql := fmt.Sprintf(sqlQuery, c.ID, c.ID, c.ID, c.ID, keypath, op, val)

	r, err := c.database.sqlite.QueryContext(ctx, sql)
	if err != nil {
		return nil, err
	}

	docs := make([]*Document, 0)
	for r.Next() {
		if err := r.Err(); err != nil {
			return nil, err
		}

		doc := Document{}

		if err := r.Scan(&doc.ID, &doc.data); err != nil {
			return nil, err
		}

		docs = append(docs, &doc)
	}

	return docs, nil
}

// Document represents a JSON document stored in a Collection.
type Document struct {
	collection *Collection
	ID         string
	data       []byte
}

// DataTo unmarshals the JSON data into the doc type if the JSON data exists.
func (d *Document) DataTo(doc any) error {
	if d.data == nil {
		return errors.New("no data")
	}

	return json.Unmarshal(d.data, &doc)
}

// Create will create a new Document with the doc type within the Collection it
// references creating the Collection if it does not already exist. The Document
// is stored as it's JSON encoded format.
func (d *Document) Create(ctx context.Context, doc any) error {
	_, err := d.collection.database.sqlite.ExecContext(ctx, fmt.Sprintf(sqlCreateTable, d.collection.ID))
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)

	err = json.NewEncoder(buf).Encode(doc)
	if err != nil {
		return err
	}

	_, err = d.collection.database.sqlite.ExecContext(ctx, fmt.Sprintf(sqlInsert, d.collection.ID, d.ID, buf.String()))
	if err != nil {
		return err
	}

	return nil
}

// Set will update a Document with the doc type within the Collection it
// references creating the Collection if it does not already exist. Set will
// fail if the Document does not already exist in the database. Create should be
// used first.
func (d *Document) Set(ctx context.Context, doc any) error {
	_, err := d.collection.database.sqlite.ExecContext(ctx, fmt.Sprintf(sqlCreateTable, d.collection.ID))
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)

	err = json.NewEncoder(buf).Encode(doc)
	if err != nil {
		return err
	}

	_, err = d.collection.database.sqlite.ExecContext(ctx, fmt.Sprintf(sqlUpdate, d.collection.ID, buf.String(), d.ID))
	if err != nil {
		return err
	}

	return nil
}

// Get will find a single Document by it's ID and call DataTo for you to
// decode the JSON into the doc's type.
func (d *Document) Get(ctx context.Context, doc any) error {
	r := d.collection.database.sqlite.QueryRowContext(ctx, fmt.Sprintf(sqlSelect, d.collection.ID, d.ID))
	if r.Err() != nil {
		return r.Err()
	}

	data := make([]byte, 0)
	err := r.Scan(&data)
	if err != nil {
		return err
	}

	d.data = data

	return d.DataTo(doc)
}

// Delete will remove the Document from the Collection it references.
func (d *Document) Delete(ctx context.Context) error {
	_, err := d.collection.database.sqlite.ExecContext(ctx, fmt.Sprintf(sqlDelete, d.collection.ID, d.ID))
	if err != nil {
		return err
	}

	return nil
}
