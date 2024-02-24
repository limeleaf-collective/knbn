package pkg

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/limeleaf-coop/knbn/pkg/db"
	"github.com/limeleaf-coop/knbn/templs"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	templ.Handler(templs.IndexPage()).ServeHTTP(w, r)
}

func BoardsHandler(db *db.Database) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		docs, err := db.Collection("boards").QueryAll(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		boards := make([]templs.Board, len(docs))
		for idx, doc := range docs {
			var board templs.Board
			if err := doc.DataTo(&board); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			board.ID = doc.ID
			boards[idx] = board
		}

		t := templs.BoardsPage(boards)
		templ.Handler(t).ServeHTTP(w, r)
	}
}

func BoardHandler(db *db.Database) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var board templs.Board
		err := db.Collection("boards").Document(r.PathValue("id")).Get(r.Context(), &board)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		t := templs.BoardPage(board)
		templ.Handler(t).ServeHTTP(w, r)
	}
}
