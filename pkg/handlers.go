package pkg

import (
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/limeleaf-coop/knbn/templs"
)

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	templ.Handler(templs.IndexPage()).ServeHTTP(w, r)
}

func BoardsHandler(w http.ResponseWriter, r *http.Request) {
	t := templs.BoardsPage(db)
	templ.Handler(t).ServeHTTP(w, r)
}

func BoardHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	board := db[id]

	t := templs.BoardPage(board)
	templ.Handler(t).ServeHTTP(w, r)
}

var db = []templs.Board{
	{
		Title: "Limeleaf Operations",
		Lists: []templs.List{
			{
				Title: "Backlog",
				Cards: []templs.Card{
					{
						Title: "Set up LLC",
						Desc:  "Still need to figure out how to LLC",
					},
				},
			},
			{
				Title: "Doing",
				Cards: []templs.Card{
					{
						Title: "Decide on Email",
						Desc:  "Do we stick with forwarding, Fastmail, or Google Workspace?",
					},
				},
			},
			{
				Title: "Done",
				Cards: []templs.Card{
					{
						Title: "Decide on Notion",
						Desc:  "Do we just pay for it and use it?",
					},
				},
			},
		},
	},
	{
		Title: "Limeleaf CRM",
		Lists: []templs.List{
			{
				Title: "Leads",
				Cards: []templs.Card{
					{
						Title: "Glens Falls School District",
						Desc:  "A bazillion dollars worth!",
					},
				},
			},
			{
				Title: "Qualified",
				Cards: []templs.Card{},
			},
			{
				Title: "Signed",
				Cards: []templs.Card{
					{
						Title: "NYS Pay Tickets",
						Desc:  "$100k",
					},
				},
			},
		},
	},
}
