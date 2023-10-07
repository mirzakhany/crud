package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mirzakhany/admin-gen/admin"
)

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
			SelectColumns: []string{"id", "name", "email", "password"},
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
