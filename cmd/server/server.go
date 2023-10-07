package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mirzakhany/admin-gen/admin"
)

// User is a sample model
type User struct {
	ID        int64     `json:"id" crud:"primary_key" crud_format:"%d"`
	Name      string    `json:"name" crud:"select,insert,update" crud_format:"%s"`
	Email     string    `json:"email" crud:"select,insert,update"`
	Password  string    `json:"password" crud:"select,insert,update"`
	UpdateAt  time.Time `json:"update_at"`
	CreatedAt time.Time `json:"created_at" crud:"select"`
}

// CrudEntity implements admin.CrudEntity interface
func (u User) CrudEntity() admin.Entity {
	return admin.Entity{
		TableName:     "users",
		TitlePlural:   "Users",
		TitleSingular: "User",
		Description:   "Users of the system.",
		FavIcon:       "fa-user",
		Order:         1,
	}
}

// Organization is a sample model
type Organization struct {
	ID        int64     `json:"id" crud:"primary_key"`
	Name      string    `json:"name" crud:"select,insert,update"`
	UpdateAt  time.Time `json:"update_at"`
	CreatedAt time.Time `json:"created_at" crud:"select"`
}

// Permission is a sample model
type Permission struct {
	ID        int64     `json:"id" crud:"primary_key"`
	Name      string    `json:"name" crud:"select,insert,update"`
	UpdateAt  time.Time `json:"update_at"`
	CreatedAt time.Time `json:"created_at" crud:"select"`
}

func main() {
	httpRouter := chi.NewRouter()
	// httpRouter.Use(middleware.Logger)
	//admin.Init(httpRouter)

	entities := []admin.Entity{
		{
			TableName:     "users",
			PrimaryKey:    "id",
			TitlePlural:   "Users",
			TitleSingular: "User",
			Description:   "Users of the system.",
			SelectColumns: []string{"id", "name", "email"},
			EditColumns:   []string{"name", "email", "password"},
			FavIcon:       "fa-user",
			Order:         1,
		},
		{
			TableName:     "organizations",
			PrimaryKey:    "id",
			TitlePlural:   "Organizations",
			TitleSingular: "Organization",
			Description:   "Organizations of the system.",
			SelectColumns: []string{"id", "name"},
			EditColumns:   []string{"name"},
			FavIcon:       "fa-building",
			Order:         2,
		},
		{
			TableName:     "permissions",
			PrimaryKey:    "id",
			TitlePlural:   "Permissions",
			TitleSingular: "Permission",
			Description:   "User permissions",
			SelectColumns: []string{"id", "name"},
			EditColumns:   []string{"name"},
			FavIcon:       "fa-key",
			Order:         3,
		},
	}

	a, err := admin.New(
		admin.WithDatabaseURI("postgres://postgres:postgres@localhost:15432/postgres?sslmode=disable"),
		admin.WithBaseURL("/admin"),
		admin.WithEntities(entities))

	if err != nil {
		panic(err)
	}

	a.PrepareHandlers(httpRouter)

	server := http.Server{
		Addr:    ":8080",
		Handler: httpRouter,
	}

	server.ListenAndServe()
}
