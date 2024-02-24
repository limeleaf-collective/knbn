package templs

type Board struct {
	ID    string
	Title string
	Lists []List
}

type List struct {
	Title string
	Cards []Card
}

type Card struct {
	Title string
	Desc  string
}
