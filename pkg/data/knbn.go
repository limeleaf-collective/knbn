package data

type Board struct {
	ID      string
	Title   string   `json:"title"`
	ListIDs []string `json:"list_ids"`
	Lists   []List
}

type List struct {
	ID      string
	Title   string   `json:"title"`
	CardIDs []string `json:"card_ids"`
	Cards   []Card
}

type Card struct {
	ID    string
	Title string `json:"title"`
	Desc  string `json:"desc"`
}
