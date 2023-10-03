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

func getTableColumenRows(db *sql.Conn, tableName string, selectColumns []string) ([]string, [][]any, error) {
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

	out := make([][]any, 0)

	for rows.Next() {
		values := make([]any, len(columns))
		for i := range columns {
			values[i] = &values[i]
		}

		if err := rows.Scan(values...); err != nil {
			return nil, nil, err
		}

		out = append(out, values)
	}

	return columns, out, nil
}
