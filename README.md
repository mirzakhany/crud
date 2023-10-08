Crud - A lightweight and user-friendly web-based admin panel for Golang applications.
========

Crud is a lightweight and user-friendly web-based admin panel for Golang applications. It simplifies the process of managing database tables by offering a straightforward CRUD (Create, Read, Update, Delete) user interface. With minimal configuration, you can quickly set up and customize an admin dashboard for your application, saving you time and effort.

Installation
------------
To install Crud, run the following command:

```bash
go get github.com/mirzakhany/crud
```

Features
--------

-   **User-friendly interface** - Crud provides a simple and intuitive user interface for managing database tables. It is designed to be easy to use, even for non-technical users.

-   **Minimal configuration** - Crud is designed to be easy to set up and use. It requires minimal configuration and can be integrated into your application in minutes.

-   **Customizable** - Crud is highly customizable. You can easily change the look and feel of the admin panel by modifying the HTML templates and CSS stylesheets.

How to use
----------
To use Crud, you need to create a new instance of the Crud struct and pass chi router to it. The following example shows how to create a new instance of the Crud struct:

```go
package main

import (
    "net/http"

    "github.com/mirzakhany/crud"
)

func main() {
   
    // Define your entities.
	entities := []crud.Entity{
		{
			TableName:     "tasks",
			PrimaryKey:    "id",
			TitlePlural:   "Tasks",
			TitleSingular: "task",
			Description:   "User tasks",
			SelectColumns: []string{"id", "name", "description", "status"},
			EditColumns:   []string{"name", "description", "status"},
			FavIcon:       "fa-tasks",
			Order:         1,
		},
	}

    // Create a new instance of the Crud struct.
   	a, err := crud.New(
		crud.WithDatabaseURI("postgres://postgres:postgres@localhost:15432/postgres?sslmode=disable"),
		crud.WithBaseURL("/admin"),
		crud.WithEntities(entities))

	if err != nil {
		panic(err)
	}

    // Create a new HTTP router.
    httpRouter := chi.NewRouter()
    // Register the Crud handlers.
	a.PrepareHandlers(httpRouter)

	server := http.Server{
		Addr:    ":8080",
		Handler: httpRouter,
	}

    // Start the HTTP server. and open http://localhost:8080/admin in your browser.
	server.ListenAndServe()
}
```

Screenshots
-----------

![Dashboard](https://raw.githubusercontent.com/mirzakhany/crud/main/screenshots/dashboard.png)
![Create](https://raw.githubusercontent.com/mirzakhany/crud/main/screenshots/create.png)
![Retrive](https://raw.githubusercontent.com/mirzakhany/crud/main/screenshots/retrive.png)
![Update](https://raw.githubusercontent.com/mirzakhany/crud/main/screenshots/update.png)
![Delete](https://raw.githubusercontent.com/mirzakhany/crud/main/screenshots/delete.png)


License
-------
This project is licensed under the terms of the MIT license.

Contributing
------------
Contributions are welcome, and they are greatly appreciated! Every little bit helps, and credit will always be given.
