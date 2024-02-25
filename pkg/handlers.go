package pkg

import (
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"
	docdb "github.com/limeleaf-coop/knbn/pkg/db"
	"github.com/limeleaf-coop/knbn/templs"
)

func metaRefresh(w http.ResponseWriter, url string) {
	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintf(w, "<meta http-equiv=\"refresh\" content=\"0; url=%s\">", url)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("knbn")
	if err != nil {
		templ.Handler(templs.IndexPage()).ServeHTTP(w, r)
		return
	}

	metaRefresh(w, "/boards")
}

func SignInHandler(db *docdb.Database) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		email := r.Form.Get("email")
		results, err := db.Collection("accounts").Query(r.Context(), "$.Email", docdb.OpEqual, email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(results) <= 0 {
			metaRefresh(w, "/")
			return
		}

		cookie := http.Cookie{
			Name:    "knbn",
			Value:   email,
			Expires: time.Now().Add(30 * time.Minute),
		}
		http.SetCookie(w, &cookie)

		metaRefresh(w, "/boards")
		return
	}
}

func BoardsHandler(db *docdb.Database) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("knbn")
		if err != nil {
			metaRefresh(w, "/")
			return
		}

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

func BoardHandler(db *docdb.Database) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("knbn")
		if err != nil {
			metaRefresh(w, "/")
			return
		}

		var board templs.Board
		err = db.Collection("boards").Document(r.PathValue("id")).Get(r.Context(), &board)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		t := templs.BoardPage(board)
		templ.Handler(t).ServeHTTP(w, r)
	}
}
