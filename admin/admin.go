package admin

import (
	"database/sql"
	"embed"
	"html/template"
	"net/http"
	"path"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

//go:embed templates
var templates embed.FS

//go:embed assets
var assets embed.FS

// handler represents a http handler.
type handler func(w http.ResponseWriter, r *http.Request)

// Formatter represents a column formatter.
type Formatter func(in any) string

// Entity represents a database entity.
type Entity struct {
	// Entity table name.
	TableName string
	// Primary key column name. default is "id".
	PrimaryKey string
	// Entity title plural. default is the entity table name.
	TitlePlural string
	// Entity title singular. default is the entity table name.
	TitleSingular string
	// FavIcon represents the entity fav icon. default is empty.
	FavIcon string
	// Menu order. default is 0.
	Order int
	// Entity description. default is empty.
	Description string
	// Entity columns. if nil, all columns will be selected.
	Columns []string
	// Formatter for each column. if provided, the formatter will be used to format the column value.
	Formatters map[string]Formatter
}

// Admin represents the admin module.
type Admin struct {
	// DatabaseURI represents the database uri. default is "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable".
	DatabaseURI string
	// DatabaseEngine represents the database engine. default is "postgres".
	databaseEngine string
	// DatabaseConn represents the database connection.
	databaseConn *sql.Conn

	// Entities represents the entities of the admin module.
	Entities map[string]Entity
	// BaseURL represents the base url for the admin module. default is "/admin".
	BaseURL string
	// DefaultFormatters represents the default formatters for each column. if provided, the formatter will be used to format the column value.
	// if a column has a formatter, the formatter will be used instead of the default formatter.
	DefaultFormatters map[string]Formatter
	// TemplateFuncs represents the template funcs for the admin module.
	TemplateFuncs template.FuncMap
	// Templates represents the templates for the admin module. will override the default templates.
	Templates map[string]*template.Template
}

// New returns a new admin module.
func New(opts ...Option) (*Admin, error) {
	a := &Admin{
		Entities: make(map[string]Entity),
	}

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}

	if a.DatabaseURI == "" {
		a.DatabaseURI = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	if a.databaseEngine == "" {
		a.databaseEngine = "postgres"
	}

	if a.BaseURL == "" {
		a.BaseURL = "/admin"
	}

	conn, err := openDB(a.databaseEngine, a.DatabaseURI)
	if err != nil {
		return nil, err
	}

	a.databaseConn = conn

	return a, nil
}

// PrepareHandlers prepares the admin module handlers.
func (a *Admin) PrepareHandlers(mux *http.ServeMux) {

	baseURL := func(s string) string {
		return path.Join(a.BaseURL, s)
	}

	// Serve static files.
	mux.Handle(baseURL("/assets/"), http.StripPrefix(baseURL("/assets/"), http.FileServer(http.FS(assets))))

	// regiser esentional handlers
	mux.HandleFunc(baseURL("/dashboard/"), a.dashboard)
	mux.HandleFunc(baseURL("/forget-password/"), a.forgetPassword)
	mux.HandleFunc(baseURL("/login/"), a.login)
	mux.HandleFunc(baseURL("/register/"), a.register)
	mux.HandleFunc(baseURL("/entity/")+"/", a.handleEntity)
}

func (a *Admin) executeTemplate(w http.ResponseWriter, name string, data any) error {
	tmpl, err := template.ParseFS(templates, "templates/*.html")
	if err != nil {
		return err
	}

	// add template funcs
	tmpl.Funcs(a.TemplateFuncs)

	// add default formatters
	for column, formatter := range a.DefaultFormatters {
		if _, ok := a.TemplateFuncs[column]; !ok {
			a.TemplateFuncs[column] = formatter
		}
	}

	// template overrides
	if tmpl, ok := a.Templates[name]; ok {
		return tmpl.Execute(w, data)
	}

	return tmpl.ExecuteTemplate(w, name, data)
}

func (a *Admin) handleEntity(w http.ResponseWriter, r *http.Request) {
	// if url contain edit then remove it and call edit handler
	editMode := false
	if strings.Contains(r.URL.Path, "/edit") {
		editMode = true
		r.URL.Path = strings.Replace(r.URL.Path, "/edit", "", 1)
	}

	// get entity name from url and call list with that name
	entityName := strings.TrimPrefix(r.URL.Path, path.Join(a.BaseURL, "/entity/")+"/")

	entity, ok := a.Entities[entityName]
	if !ok {
		http.Error(w, "entity not found", http.StatusNotFound)
		return
	}

	if r.Method == http.MethodGet {
		if editMode {
			a.handleGetEdit(w, r, entity)
		} else {
			a.handleListGet(w, r, entity)
		}
	} else if r.Method == http.MethodPost {
	}
}

// ListData represents the data needed to render the list template.
type ListData struct {
	Title       string
	Description string
	Columns     []string
	Rows        [][]any

	Menus []Menu
}

func (a *Admin) handleListGet(w http.ResponseWriter, r *http.Request, entity Entity) {
	cols, rows, err := getTableColumenRows(a.databaseConn, entity.TableName, entity.Columns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	data := ListData{
		Menus:       a.getMenus(),
		Title:       entity.TitlePlural,
		Description: entity.Description,
		Columns:     cols,
		Rows:        rows,
	}

	if err := a.executeTemplate(w, "list", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type EditData struct {
	Title       string
	Description string
	Columns     []string

	ColumnData map[string]any

	Menus []Menu
}

func (a *Admin) handleGetEdit(w http.ResponseWriter, r *http.Request, entity Entity) {
	data := EditData{
		Menus:       a.getMenus(),
		Title:       entity.TitleSingular,
		Description: entity.Description,
		Columns:     entity.Columns,
	}

	if err := a.executeTemplate(w, "edit", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Menu represents a menu item.
type Menu struct {
	Order     int
	Idenifier string
	Title     string
	URL       string
	FavIcon   string
}

func (a *Admin) getMenus() []Menu {
	menus := make([]Menu, 0)

	for _, entity := range a.Entities {
		menus = append(menus, Menu{
			Idenifier: entity.TableName,
			Title:     entity.TitlePlural,
			URL:       path.Join(a.BaseURL, "/entity/", entity.TableName),
			FavIcon:   entity.FavIcon,
			Order:     entity.Order,
		})
	}

	// sort menus by order
	sort.Slice(menus, func(i, j int) bool {
		return menus[i].Order < menus[j].Order
	})

	return menus
}

func (a *Admin) dashboard(w http.ResponseWriter, r *http.Request) {
	if err := a.executeTemplate(w, "dashboard", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *Admin) login(w http.ResponseWriter, r *http.Request) {
	if err := a.executeTemplate(w, "login", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *Admin) register(w http.ResponseWriter, r *http.Request) {
	if err := a.executeTemplate(w, "register", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *Admin) forgetPassword(w http.ResponseWriter, r *http.Request) {
	if err := a.executeTemplate(w, "forget_password", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
