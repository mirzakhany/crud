package admin

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// DB represents a database.
type DB struct {
	URI    string
	Engine string
}

// Open opens a database connection.
func (d *DB) Open(ctx context.Context) (*sql.Conn, error) {
	db, err := sql.Open(d.Engine, d.URI)
	if err != nil {
		return nil, err
	}
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db.Conn(ctx)
}

// Row represents a row of a table.
type Row struct {
	Columns         []Column
	PrimaryKey      string
	PrimaryKeyValue any
}

// Column represents a column of a table.
type Column struct {
	Name      string
	Type      string
	Value     any
	IsPrimary bool
}

// GetTableColumenRows returns the rows of a table.
func (d *DB) GetTableColumenRows(ctx context.Context, tableName, primaryKey string, selectColumns []string) ([]Row, []string, error) {
	if len(selectColumns) == 0 {
		selectColumns = []string{"*"}
	}

	db, err := d.Open(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	stmt := fmt.Sprintf("select %s from %s", strings.Join(selectColumns, ","), tableName)
	rows, err := db.QueryContext(ctx, stmt)

	if err != nil {
		return nil, nil, err
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, err
	}

	out := make([]Row, 0)

	for rows.Next() {
		values := make([]any, len(columns))
		for i := range columns {
			values[i] = &values[i]
		}

		if err := rows.Scan(values...); err != nil {
			return nil, nil, err
		}

		row := Row{
			PrimaryKey: primaryKey,
			Columns:    make([]Column, 0),
		}

		for i, column := range columns {
			row.Columns = append(row.Columns, Column{
				Name:      column,
				Type:      fieldTypeToGo(columnTypes[i].DatabaseTypeName()),
				Value:     values[i],
				IsPrimary: column == primaryKey,
			})

			if column == primaryKey {
				row.PrimaryKeyValue = values[i]
			}
		}

		out = append(out, row)
	}

	return out, columns, nil
}

// GetEntityByID returns a row of a table by its primary key.
func (d *DB) GetEntityByID(ctx context.Context, tableName, primaryKey string, editColumns []string, id any) (*Row, error) {
	db, err := d.Open(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	stmt := fmt.Sprintf("select %s from %s where %s = $1 limit 1", strings.Join(editColumns, ","), tableName, primaryKey)
	rows, err := db.QueryContext(ctx, stmt, id)
	if err != nil {
		return nil, err
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	values := make([]any, len(columns))
	for i := range columns {
		values[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
	}

	out := &Row{
		PrimaryKey: primaryKey,
	}

	cols := make([]Column, 0)
	for i, column := range columns {
		cols = append(cols, Column{
			Name:      column,
			Type:      fieldTypeToGo(columnTypes[i].DatabaseTypeName()),
			Value:     values[i],
			IsPrimary: column == primaryKey,
		})

		if column == primaryKey {
			out.PrimaryKeyValue = values[i]
		}
	}

	out.Columns = cols

	return out, nil
}

// DeleteEntityByID deletes a row of a table by its primary key.
func (d *DB) DeleteEntityByID(ctx context.Context, tableName, primaryKey string, id any) error {
	db, err := d.Open(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	stmt := fmt.Sprintf("delete from %s where %s = $1", tableName, primaryKey)
	if _, err := db.ExecContext(ctx, stmt, id); err != nil {
		return err
	}

	return nil
}

// CreateEntity creates a row of a table.
func (d *DB) CreateEntity(ctx context.Context, tableName, primaryKey string, columns []Column) error {
	db, err := d.Open(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	cols := make([]string, 0)
	values := make([]any, 0)

	for _, column := range columns {
		if column.Name == primaryKey {
			continue
		}

		cols = append(cols, column.Name)
		values = append(values, column.Value)
	}

	stmt := fmt.Sprintf("insert into %s (%s) values (%s)", tableName, strings.Join(cols, ","), strings.Join(strings.Split(strings.Repeat("?", len(cols)), ""), ","))
	if _, err := db.ExecContext(ctx, stmt, values...); err != nil {
		return err
	}

	return nil
}

// GetTableRow returns the columns of a table.
func (d *DB) GetTableRow(ctx context.Context, tableName, primaryKey string, editColumns []string) (*Row, error) {
	db, err := d.Open(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if len(editColumns) == 0 {
		editColumns = []string{"*"}
	}

	rows, err := db.QueryContext(ctx, fmt.Sprintf("select %s from %s limit 0", strings.Join(editColumns, ","), tableName))
	if err != nil {
		return nil, err
	}

	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	// dummy scan
	values := make([]any, len(columns))
	for i := range columns {
		values[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
	}
	// end dummy scan

	out := &Row{
		PrimaryKey: primaryKey,
	}

	for _, column := range columns {
		out.Columns = append(out.Columns, Column{
			Name:      column.Name(),
			Type:      fieldTypeToGo(column.DatabaseTypeName()),
			IsPrimary: column.Name() == primaryKey,
		})
	}

	return out, nil
}

// GetTableFieldTypes returns the field types of a table.
func (d *DB) GetTableFieldTypes(ctx context.Context, tableName string) (map[string]string, error) {
	db, err := d.Open(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, fmt.Sprintf("select * from %s limit 0", tableName))
	if err != nil {
		return nil, err
	}

	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	out := make(map[string]string)

	for _, column := range columns {
		out[column.Name()] = column.DatabaseTypeName()
	}

	return out, nil
}

func fieldTypeToGo(fieldType string) string {
	switch fieldType {
	case "int", "int2", "int4", "int8", "smallint", "integer", "bigint", "INT4":
		return "int"
	case "float", "float4", "float8", "decimal", "numeric", "real", "double precision":
		return "float64"
	case "bool", "boolean":
		return "bool"
	case "date", "timestamp", "timestamp with time zone", "timestamp without time zone", "TIMESTAMP":
		return "time.Time"
	case "text", "varchar", "character varying", "character", "TEXT":
		return "string"
	default:
		fmt.Println("unknown type", fieldType)
		return "any"
	}
}

func goTypeToFieldType(goType string) string {
	switch goType {
	case "int":
		return "int"
	case "float64":
		return "float"
	case "bool":
		return "bool"
	case "time.Time":
		return "timestamp with time zone"
	case "string":
		return "text"
	default:
		return "any"
	}
}
