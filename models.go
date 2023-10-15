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

// SearchResult represents the search results.
type SearchResult struct {
	Title       string
	Description string
	Link        string
}

// SearchData represents the data needed to render the search template.
type SearchData struct {
	Query       string
	ResultCount int

	BaseContextData

	Results []SearchResult
}

// BaseContextData represents the data needed to render the base template.
type BaseContextData struct {
	ShowSearchBar bool
	BaseURL       string
	Menus         []Menu
}
