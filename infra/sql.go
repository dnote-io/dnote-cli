package infra

import (
	"database/sql"

	"github.com/pkg/errors"
)

// SQLCommon is the minimal interface required by a db connection
type SQLCommon interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// sqlDb is an interface implemented by *sql.DB
type sqlDb interface {
	Begin() (*sql.Tx, error)
}

// sqlTx is an interface implemented by *sql.Tx
type sqlTx interface {
	Commit() error
	Rollback() error
}

// DB contains information about the current database connection
type DB struct {
	Conn SQLCommon
}

// OpenDB initializes a new connection to the sqlite database
func OpenDB(dbPath string) (*DB, error) {
	dbConn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "opening db connection")
	}

	// Send a ping to ensure that the connection is established
	if err := dbConn.Ping(); err != nil {
		dbConn.Close()
		return nil, errors.Wrap(err, "sending a ping")
	}

	db := &DB{
		Conn: dbConn,
	}

	return db, nil
}

// Begin begins a transaction
func (d *DB) Begin() (*DB, error) {
	if db, ok := d.Conn.(sqlDb); ok && db != nil {
		tx, err := db.Begin()
		if err != nil {
			return nil, err
		}

		return &DB{Conn: tx}, nil
	}

	return nil, errors.New("can't start transaction")
}

// Commit commits a transaction
func (d *DB) Commit() error {
	if db, ok := d.Conn.(sqlTx); ok && db != nil {
		if err := db.Commit(); err != nil {
			return err
		}
	}

	return errors.New("invalid transaction")
}

// Rollback rolls back a transaction
func (d *DB) Rollback() error {
	if db, ok := d.Conn.(sqlTx); ok && db != nil {
		if err := db.Rollback(); err != nil {
			return err
		}
	}

	return errors.New("invalid transaction")
}

// Exec executes a sql
func (d *DB) Exec(query string, values ...interface{}) (sql.Result, error) {
	return d.Conn.Exec(query, values...)
}

// Prepare prepares a sql
func (d *DB) Prepare(query string) (*sql.Stmt, error) {
	return d.Conn.Prepare(query)
}

// Query queries rows
func (d *DB) Query(query string, values ...interface{}) (*sql.Rows, error) {
	return d.Conn.Query(query, values...)
}

// QueryRow queries a row
func (d *DB) QueryRow(query string, values ...interface{}) *sql.Row {
	return d.Conn.QueryRow(query, values...)
}

type closer interface {
	Close() error
}

// Close closes a db connection
func (d *DB) Close() error {
	if db, ok := d.Conn.(closer); ok {
		return db.Close()
	}

	return errors.New("can't close db")
}
