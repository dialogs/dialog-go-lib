package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
)

type TestEnv struct {
	conn *sql.DB
}

func NewTestEnv(conn string) (*TestEnv, error) {

	conf, err := pgx.ParseConfig(conn)
	if err != nil {
		return nil, err
	}

	return &TestEnv{
		conn: stdlib.OpenDB(*conf),
	}, nil
}

func (c *TestEnv) Close() error {
	return c.conn.Close()
}

func (c *TestEnv) Conn() *sql.DB {
	return c.conn
}

func (c *TestEnv) ClearDB(queries ...string) (err error) {

	var tx *sql.Tx
	tx, err = c.conn.BeginTx(context.Background(), nil)
	if err != nil {
		return
	}

	defer func() {
		if err == nil {
			err = tx.Commit()

		} else {
			if opErr := tx.Rollback(); opErr != nil {
				log.Println("ERROR:", opErr)
			}

		}
	}()

	for _, q := range queries {
		if _, err = tx.Exec(q); err != nil {
			return
		}
	}

	return
}

func (c *TestEnv) CreateScheme(schema string) (err error) {

	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema)
	_, err = c.conn.Exec(query)
	return
}

func (c *TestEnv) DropScheme(schema string) (err error) {

	query := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema)
	_, err = c.conn.Exec(query)
	return
}

func (c *TestEnv) SchemeExists(schema string) (err error) {

	const Query = "SELECT true FROM information_schema.schemata WHERE schema_name = $1"

	var exists bool
	err = c.conn.QueryRow(Query, schema).Scan(&exists)
	if err == nil && !exists {
		err = sql.ErrNoRows
	}

	return
}

func (c *TestEnv) GetTableColumns(scheme, table string) (map[string]string, error) {

	const Query = `
	SELECT column_name, data_type
	FROM information_schema.columns
	WHERE table_schema = $1 AND table_name = $2
	ORDER BY column_name`

	rows, err := c.conn.Query(Query, scheme, table)
	if err != nil {
		return nil, err
	}

	defer func() {
		if opErr := rows.Close(); opErr != nil {
			log.Println("ERROR:", opErr)
		}
	}()

	var (
		columnName string
		dataType   string
		columns    = make(map[string]string)
	)

	for rows.Next() {
		if err := rows.Scan(&columnName, &dataType); err != nil {
			return nil, err
		}

		columns[columnName] = dataType
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}
