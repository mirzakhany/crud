package admin

import (
	"context"
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
	// SelectColumns columns. if nil, all columns will be selected.
	SelectColumns []string
	// EditColumns columns. if nil, all columns will be selected.
	EditColumns []string
	// NewColumns columns. if nil, edit columns or all columns will be selected.
	NewColumns []string
	// ColumnNameFormatter represents the column name formatter. if provided, the formatter will be used to format the column name.
	ColumnNameFormatter map[string]Formatter
	// ValueFormatters for each column. if provided, the formatter will be used to format the column value.
	ValueFormatters map[string]Formatter
}

// Admin represents the admin module.
type Admin struct {
	// DatabaseURI represents the database uri. default is "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable".
	DatabaseURI string
	// DatabaseEngine represents the database engine. default is "postgres".
	databaseEngine string
	// DatabaseConn represents the database connection.
	db *DB
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

	a.db = &DB{
		URI:    a.DatabaseURI,
		Engine: a.databaseEngine,
	}

	if _, err := a.db.Open(context.Background()); err != nil {
		return nil, err
	}
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

	r.Get(baseURL("/"), a.dashboard)
	r.Get(baseURL("/search/"), a.searchView)
	r.Get(baseURL("/entity/{entity}"), a.getEntityList)
	r.Get(baseURL("/entity/{entity}/new"), a.getEntityNew)
	r.Post(baseURL("/entity/{entity}/new"), a.createEntity)
	r.Get(baseURL("/entity/{entity}/{entityID}"), a.getEntityEdit)
	r.Post(baseURL("/entity/{entity}/{entityID}"), a.updateEntity)
	r.Get(baseURL("/entity/{entity}/{entityID}/delete"), a.deleteEntity)
}

// ListData represents the data needed to render the list template.
type ListData struct {
	Title       string
	Description string
	EntityName  string

	Columns []string
	Rows    []Row

	BaseURL string
	Menus   []Menu
}

type SearchResults struct {
	Title string
	Link  string
}

type SearchData struct {
	Query       string
	ResultCount int

	BaseURL string
	Menus   []Menu

	Results []SearchResults
}

func (a *Admin) searchView(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	data := SearchData{
		BaseURL: a.BaseURL,
		Menus:   a.getMenus(),

		Query:       query,
		ResultCount: 2,
		Results: []SearchResults{
			{
				Title: "test",
				Link:  "test",
			},
			{
				Title: "test2",
				Link:  "test2",
			},
		},
	}

	if err := a.executeTemplate(w, "search", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *Admin) createEntity(w http.ResponseWriter, r *http.Request) {
	entityName := chi.URLParam(r, "entity")
	entity, ok := a.Entities[entityName]
	if !ok {
		a.renderNotFoundPage(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	columns := make([]Column, 0)

	for column, value := range r.Form {
		if column == entity.PrimaryKey {
			continue
		}

		columns = append(columns, Column{
			Name:  column,
			Value: value[0],
		})
	}

	if err := a.db.CreateEntity(r.Context(), entity.TableName, entity.PrimaryKey, columns); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, path.Join(a.BaseURL, "/entity/", entity.TableName), http.StatusFound)
}

func (a *Admin) updateEntity(w http.ResponseWriter, r *http.Request) {
	entityName := chi.URLParam(r, "entity")
	entityID := chi.URLParam(r, "entityID")

	entity, ok := a.Entities[entityName]
	if !ok {
		a.renderNotFoundPage(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	columns := make([]Column, 0)
	for column, value := range r.Form {
		if column == entity.PrimaryKey {
			continue
		}

		columns = append(columns, Column{
			Name:  column,
			Value: value[0],
		})
	}

	if err := a.db.UpdateEntity(r.Context(), entity.TableName, entity.PrimaryKey, entityID, columns); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, path.Join(a.BaseURL, "/entity/", entity.TableName, entityID), http.StatusFound)
}

func (a *Admin) getEntityList(w http.ResponseWriter, r *http.Request) {
	entityName := chi.URLParam(r, "entity")
	entity, ok := a.Entities[entityName]
	if !ok {
		a.renderNotFoundPage(w, r)
		return
	}

	rows, columens, err := a.db.GetTableColumenRows(r.Context(), entity.TableName, entity.PrimaryKey, entity.getSelectColumns())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	data := ListData{
		Menus:       a.getMenus(),
		Title:       entity.TitlePlural,
		EntityName:  entity.TableName,
		BaseURL:     a.BaseURL,
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
	EntityName  string
	EntityID    string

	IsEdit bool

	BaseURL string
	Menus   []Menu
}

func (a *Admin) getEntityEdit(w http.ResponseWriter, r *http.Request) {
	// get entity name from url and call list with that name
	entityName := chi.URLParam(r, "entity")
	entityID := chi.URLParam(r, "entityID")

	entity, ok := a.Entities[entityName]
	if !ok {
		a.renderNotFoundPage(w, r)
		return
	}

	row, err := a.db.GetEntityByID(r.Context(), entity.TableName, entity.PrimaryKey, entity.getEditColumns(), entityID)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err == sql.ErrNoRows {
		a.renderNotFoundPage(w, r)
		return
	}

	data := EditData{
		BaseURL:     a.BaseURL,
		Menus:       a.getMenus(),
		Title:       entity.TitleSingular,
		Description: entity.Description,
		EntityName:  entityName,
		EntityID:    entityID,

		Row:    *row,
		IsEdit: true,
	}

	if err := a.executeTemplate(w, "edit", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *Admin) getEntityNew(w http.ResponseWriter, r *http.Request) {
	// get entity name from url and call list with that name
	entityName := chi.URLParam(r, "entity")

	entity, ok := a.Entities[entityName]
	if !ok {
		a.renderNotFoundPage(w, r)
		return
	}

	row, err := a.db.GetTableRow(r.Context(), entity.TableName, entity.PrimaryKey, entity.getNewColumns())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := EditData{
		BaseURL:     a.BaseURL,
		Menus:       a.getMenus(),
		EntityName:  entityName,
		Title:       entity.TitleSingular,
		Description: entity.Description,
		Row:         *row,
		IsEdit:      false,
	}

	if err := a.executeTemplate(w, "new", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *Admin) deleteEntity(w http.ResponseWriter, r *http.Request) {
	// get entity name from url and call list with that name
	entityName := chi.URLParam(r, "entity")
	entityID := chi.URLParam(r, "entityID")

	entity, ok := a.Entities[entityName]
	if !ok {
		a.renderNotFoundPage(w, r)
		return
	}

	if err := a.db.DeleteEntityByID(r.Context(), entity.TableName, entity.PrimaryKey, entityID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, path.Join(a.BaseURL, "/entity/", entity.TableName), http.StatusFound)
}

// DashboardData represents the data needed to render the dashboard template.
type DashboardData struct {
	BaseURL string
	Menus   []Menu
}

func (a *Admin) dashboard(w http.ResponseWriter, r *http.Request) {
	dashboardData := DashboardData{
		BaseURL: a.BaseURL,
		Menus:   a.getMenus(),
	}

	if err := a.executeTemplate(w, "dashboard", dashboardData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type NotfoundPageDaa struct {
	BaseURL string
	Menus   []Menu
}

func (a *Admin) renderNotFoundPage(w http.ResponseWriter, r *http.Request) {
	data := NotfoundPageDaa{
		BaseURL: a.BaseURL,
		Menus:   a.getMenus(),
	}

	if err := a.executeTemplate(w, "not_found", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

func (e Entity) getSelectColumns() []string {
	if len(e.SelectColumns) == 0 {
		return []string{"*"}
	}

	return e.SelectColumns
}

func (e Entity) getEditColumns() []string {
	if len(e.EditColumns) == 0 {
		return []string{"*"}
	}

	return e.EditColumns
}

func (e Entity) getNewColumns() []string {
	if len(e.NewColumns) == 0 {
		if len(e.EditColumns) != 0 {
			return e.EditColumns
		}
		return []string{"*"}
	}

	return e.NewColumns
}
