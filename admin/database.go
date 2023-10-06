package admin

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func openDB(engine, uri string) (*sql.Conn, error) {
	db, err := sql.Open(engine, uri)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db.Conn(context.Background())
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

func getTableColumenRows(db *sql.Conn, tableName, primaryKey string, selectColumns []string) ([]Row, []string, error) {
	if len(selectColumns) == 0 {
		selectColumns = []string{"*"}
	}

	stmt := fmt.Sprintf("select %s from %s", strings.Join(selectColumns, ","), tableName)
	rows, err := db.QueryContext(context.Background(), stmt)
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

func getEntityByID(db *sql.Conn, tableName, primaryKey string, id any) (*Row, error) {
	stmt := fmt.Sprintf("select * from %s where %s = $1 limit 1", tableName, primaryKey)
	rows, err := db.QueryContext(context.Background(), stmt, id)
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

func getTableFieldTypes(db *sql.Conn, tableName string) (map[string]string, error) {
	rows, err := db.QueryContext(context.Background(), fmt.Sprintf("select * from %s limit 0", tableName))
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
