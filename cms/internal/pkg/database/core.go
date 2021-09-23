package database

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

// Init takes the details from config and initializes a database connection,
// bootstrap can be set to overide the database name in the params to postgres
// when first connecting to the database server
func Init(
	connectionStringBase string,
	rawParams map[string]string,
	databaseName string,
	bootstrap bool,
) (*sql.DB, error) {
	// convert the map[string]string from the config into url params for the
	// connection string
	params := url.Values{}
	for k, v := range rawParams {
		params.Add(k, v)
	}

	// if the database in question might be missing, we need to connect to the
	// server on the postgres database to create the database
	if bootstrap {
		params.Del("dbname")
		params.Add("dbname", databaseName)
	}

	// generate the final connectionString based on the params
	connectionString := fmt.Sprintf(
		"%s?%s",
		connectionStringBase,
		params.Encode())

	// Open the connection and test that it's working
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return db, fmt.Errorf("failed to init db connection: %s", err)
	}
	if err = db.Ping(); err != nil {
		return db, fmt.Errorf("failed to ping the database: %s", err)
	}

	return db, nil
}

// Create will attempt to create a new database with a given name
func Create(db *sql.DB, databaseName string) error {
	_, err := db.Exec(fmt.Sprintf(`CREATE DATABASE %s;`, databaseName))
	if err != nil {
		return fmt.Errorf("failed to create database: %s", err)
	}

	return nil
}

// Drop will terminate connections to the database and remove it
func Drop(db *sql.DB, databaseName string) error {
	// https://stackoverflow.com/questions/5408156/how-to-drop-a-postgresql-database-if-there-are-active-connections-to-it
	transactionSQL := fmt.Sprintf(`
      UPDATE pg_database SET datallowconn = 'false' WHERE datname = '%s';

      SELECT pg_terminate_backend(pid)
      FROM pg_stat_activity
      WHERE datname = '%s';`, databaseName, databaseName)

	_, err := db.Exec(transactionSQL)
	if err != nil {
		return fmt.Errorf("failed to terminate active connections: %s", err)
	}

	// once the connections have been removed, then we can drop the database
	_, err = db.Exec(fmt.Sprintf(`DROP DATABASE %s;`, databaseName))
	if err != nil {
		return fmt.Errorf("failed to drop database %q: %s", databaseName, err)
	}

	return nil
}

// Ping calls db.Ping on the connection handle to test the connection, a simple
// wrapper
func Ping(db *sql.DB) error {
	err := db.Ping()
	if err != nil {
		return errors.Wrap(err, "failed to ping database")
	}

	return nil
}

// Exists will return true if a database with the supplied name exists on the
// currently connected postgres instance
func Exists(db *sql.DB, databaseName string) (bool, error) {
	// https://dba.stackexchange.com/questions/45143/check-if-postgresql-database-exists-case-insensitive-way
	rows, err := db.Query(`SELECT 1 FROM pg_database WHERE datname=$1`, databaseName)
	if err != nil {
		return false, fmt.Errorf("failed look up database: %s", err)
	}
	defer rows.Close()

	var result int
	for rows.Next() {
		err = rows.Scan(&result)
		if err != nil {
			return false, fmt.Errorf(
				"failed to parse sql row response containing matching databases: %s",
				err,
			)
		}
	}

	// only return true if there was a matching row with 1 set for a name match
	return result == 1, nil
}

// Truncate the table with tableName
func Truncate(db *sql.DB, tableName string) error {
	_, err := db.Exec(fmt.Sprintf(`TRUNCATE %s;`, tableName))
	if err != nil {
		return fmt.Errorf("failed to truncate table: %s", err)
	}

	return nil
}
