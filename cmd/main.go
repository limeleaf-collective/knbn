package main

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/limeleaf-coop/knbn/pkg"
	"github.com/limeleaf-coop/knbn/templs"
)

func main() {
	board := pkg.Board{
		Title: "Limeleaf Operations",
		Lists: []pkg.List{
			{
				Title: "Backlog",
				Cards: []pkg.Card{
					{
						Title: "Set up LLC",
						Desc:  "Still need to figure out how to LLC",
					},
				},
			},
			{
				Title: "Doing",
				Cards: []pkg.Card{
					{
						Title: "Decide on Email",
						Desc:  "Do we stick with forwarding, Fastmail, or Google Workspace?",
					},
				},
			},
			{
				Title: "Done",
				Cards: []pkg.Card{
					{
						Title: "Decide on Notion",
						Desc:  "Do we just pay for it and use it?",
					},
				},
			},
		},
	}

	http.Handle("/boards/1234", templ.Handler(templs.BoardPage(board)))
	http.Handle("/boards", templ.Handler(templs.BoardsPage()))
	http.Handle("/", templ.Handler(templs.IndexPage()))
	http.ListenAndServe(":8080", nil)
}
