package crud

import (
	"html/template"
	"net/http"
	"path"
)

// Option represents an admin option.
type Option func(*Admin) error

// WithDatabaseURI returns an admin option that sets the database uri.
func WithDatabaseURI(uri string) Option {
	return func(a *Admin) error {
		a.DatabaseURI = uri
		a.databaseEngine = path.Ext(uri)
		return nil
	}
}

// WithBaseURL returns an admin option that sets the admin base url.
func WithBaseURL(baseURL string) Option {
	return func(a *Admin) error {
		a.BaseURL = baseURL
		return nil
	}
}

// WithDefaultFormatters returns an admin option that sets the default formatters.
func WithDefaultFormatters(formatters map[string]Formatter) Option {
	return func(a *Admin) error {
		a.DefaultFormatters = formatters
		return nil
	}
}

// WithDefaultFormatter returns an admin option that sets a default formatter.
func WithDefaultFormatter(column string, formatter Formatter) Option {
	return func(a *Admin) error {
		a.DefaultFormatters[column] = formatter
		return nil
	}
}

// WithEntity returns an admin option that adds an entity.
func WithEntity(entity Entity) Option {
	return func(a *Admin) error {
		a.Entities[entity.TableName] = entity
		return nil
	}
}

// WithEntities returns an admin option that adds entities.
func WithEntities(entities []Entity) Option {
	return func(a *Admin) error {
		for _, entity := range entities {
			a.Entities[entity.TableName] = entity
		}
		return nil
	}
}

// WithTemplateFuncs returns an admin option that adds template funcs.
func WithTemplateFuncs(funcs template.FuncMap) Option {
	return func(a *Admin) error {
		a.TemplateFuncs = funcs
		return nil
	}
}

// WithTemplateFunc returns an admin option that adds a template func.
func WithTemplateFunc(name string, fn interface{}) Option {
	return func(a *Admin) error {
		a.TemplateFuncs[name] = fn
		return nil
	}
}

// WithTemplate returns an admin option that adds a template.
func WithTemplate(name, text string) Option {
	return func(a *Admin) error {
		tmpl, err := template.New(name).Parse(text)
		if err != nil {
			return err
		}
		a.Templates[name] = tmpl
		return nil
	}
}

// WithUserIdentifier returns an admin option that sets the user identifier.
func WithUserIdentifier(fn func(r *http.Request) string) Option {
	return func(a *Admin) error {
		a.UserIdentifier = fn
		return nil
	}
}

// WithPermissionChecker returns an admin option that sets the permission checker.
func WithPermissionChecker(fn func(r *http.Request, userID, entityName, action string) bool) Option {
	return func(a *Admin) error {
		a.PermissionChecker = fn
		return nil
	}
}
