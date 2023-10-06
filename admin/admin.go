package admin

import (
	"database/sql"
	"embed"
	"html/template"
	"net/http"
	"path"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

//go:embed all:templates
var templates embed.FS

//go:embed all:assets
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
func (a *Admin) PrepareHandlers(r chi.Router) {
	baseURL := func(s string) string {
		return path.Join(a.BaseURL, s)
	}

	FileServer(r, baseURL("/"), http.FS(assets))

	// regiser esentional handlers
	// mux.HandleFunc(baseURL("/dashboard/"), a.dashboard)
	// mux.HandleFunc(baseURL("/forget-password/"), a.forgetPassword)
	// mux.HandleFunc(baseURL("/login/"), a.login)
	// mux.HandleFunc(baseURL("/register/"), a.register)
	// mux.HandleFunc(baseURL("/entity/")+"/", a.handleEntity)

	r.Get(baseURL("/entity/{entity}"), a.getEntityList)
	r.Get(baseURL("/entity/{entity}/{entityID}"), a.getEntityEdit)
	r.Get(baseURL("/entity/{entity}/{entityID}/delete"), a.deleteEntity)
}

// ListData represents the data needed to render the list template.
type ListData struct {
	Title       string
	Description string
	EntityName  string
	BaesURL     string
	Columns     []string
	Rows        []Row

	Menus []Menu
}

func (a *Admin) getEntityList(w http.ResponseWriter, r *http.Request) {
	entityName := chi.URLParam(r, "entity")
	entity, ok := a.Entities[entityName]
	if !ok {
		http.Error(w, "entity not found", http.StatusNotFound)
		return
	}

	rows, columens, err := getTableColumenRows(a.databaseConn, entity.TableName, entity.PrimaryKey, entity.Columns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	data := ListData{
		Menus:       a.getMenus(),
		Title:       entity.TitlePlural,
		EntityName:  entity.TableName,
		BaesURL:     a.BaseURL,
		Description: entity.Description,
		Columns:     columens,
		Rows:        rows,
	}

	if err := a.executeTemplate(w, "list", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// EditData represents the data needed to render the edit template.
type EditData struct {
	Title       string
	Description string
	Row         Row

	Menus []Menu
}

func (a *Admin) getEntityEdit(w http.ResponseWriter, r *http.Request) {
	// get entity name from url and call list with that name
	entityName := chi.URLParam(r, "entity")
	entityID := chi.URLParam(r, "entityID")

	entity, ok := a.Entities[entityName]
	if !ok {
		http.Error(w, "entity not found", http.StatusNotFound)
		return
	}

	row, err := getEntityByID(a.databaseConn, entity.TableName, entity.PrimaryKey, entityID)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err == sql.ErrNoRows {
		http.Error(w, "entity not found", http.StatusNotFound)
		return
	}

	data := EditData{
		Menus:       a.getMenus(),
		Title:       entity.TitleSingular,
		Description: entity.Description,
		Row:         *row,
	}

	if err := a.executeTemplate(w, "edit", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *Admin) deleteEntity(w http.ResponseWriter, r *http.Request) {
	// get entity name from url and call list with that name
	entityName := chi.URLParam(r, "entity")
	entityID := chi.URLParam(r, "entityID")

	entity, ok := a.Entities[entityName]
	if !ok {
		http.Error(w, "entity not found", http.StatusNotFound)
		return
	}

	if err := deleteEntityByID(a.databaseConn, entity.TableName, entity.PrimaryKey, entityID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, path.Join(a.BaseURL, "/entity/", entity.TableName), http.StatusFound)
}

func (a *Admin) executeTemplate(w http.ResponseWriter, name string, data any) error {
	tmpl, err := template.New("base").Funcs(templateFuncs()).ParseFS(templates, "templates/*.html")
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

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")

		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
