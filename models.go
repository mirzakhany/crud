package crud

// Menu represents a menu item.
type Menu struct {
	Order     int
	Idenifier string
	Title     string
	URL       string
	FavIcon   string
}

// EditData represents the data needed to render the edit template.
type EditData struct {
	Title       string
	Description string
	Row         Row
	EntityName  string
	EntityID    string

	IsEdit bool

	BaseContextData
}

// ListData represents the data needed to render the list template.
type ListData struct {
	Title       string
	Description string
	EntityName  string

	Columns []string
	Rows    []Row

	BaseContextData
}

// SearchResults represents the search results.
type SearchResults struct {
	Title string
	Link  string
}

// SearchData represents the data needed to render the search template.
type SearchData struct {
	Query       string
	ResultCount int

	BaseContextData

	Results []SearchResults
}

type BaseContextData struct {
	BaseURL string
	Menus   []Menu
}
