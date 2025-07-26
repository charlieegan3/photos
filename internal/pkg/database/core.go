package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type SelectOptions struct {
	SortField      string
	SortDescending bool
	Offset         uint
	Limit          uint
}

// Init takes the details from config and initializes a database connection,
// bootstrap can be set to overide the database name in the params to postgres
// when first connecting to the database server.
func Init(
	ctx context.Context,
	connectionStringBase string,
	rawParams map[string]string,
	databaseName string,
	bootstrap bool,
) (*sql.DB, error) {
	// if the database in question might be missing, we need to connect to the
	// server on the postgres database to create the database
	if bootstrap {
		params := url.Values{}
		for k, v := range rawParams {
			params.Add(k, v)
		}

		params.Del("dbname")
		params.Add("dbname", "postgres")

		// generate the final connectionString based on the params
		connectionString := buildConnectionString(connectionStringBase, params)

		// Open the connection and test that it's working
		db, err := sql.Open("postgres", connectionString)
		if err != nil {
			return db, fmt.Errorf("failed to init db connection: %w", err)
		}

		exists, err := Exists(ctx, db, databaseName)
		if err != nil {
			return db, fmt.Errorf("failed to check if database must be created: %w", err)
		}

		if !exists {
			err = Create(ctx, db, databaseName)
			if err != nil {
				return db, fmt.Errorf("failed to create database: %w", err)
			}
		}
	}

	// convert the map[string]string from the config into url params for the
	// connection string
	params := url.Values{}
	for k, v := range rawParams {
		// we should use the name of the database set in the function args
		// here, this allows us to overide the dbname if set. This can be used
		// to return a handle to the postgres database to allow dropping of the
		// test database during test runs.
		if k == "dbname" {
			params.Add(k, databaseName)
			continue
		}
		params.Add(k, v)
	}

	// generate the final connectionString based on the params
	connectionString := buildConnectionString(connectionStringBase, params)

	// Open the connection and test that it's working
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return db, fmt.Errorf("failed to init db connection: %w", err)
	}

	if err = db.PingContext(ctx); err != nil {
		return db, fmt.Errorf("failed to ping the database: %w", err)
	}

	// this is the limit set by elephantsql.com free tier, so no point in making configurable right now
	db.SetMaxOpenConns(5)

	return db, nil
}

// Create will attempt to create a new database with a given name.
func Create(ctx context.Context, db *sql.DB, databaseName string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE %s;`, databaseName))
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	return nil
}

// Drop will terminate connections to the database and remove it.
func Drop(ctx context.Context, db *sql.DB, databaseName string) error {
	// https://stackoverflow.com/questions/5408156/how-to-drop-a-postgresql-database-if-there-are-active-connections-to-it
	transactionSQL := fmt.Sprintf(`
      UPDATE pg_database SET datallowconn = 'false' WHERE datname = '%s';

      SELECT pg_terminate_backend(pid)
      FROM pg_stat_activity
      WHERE datname = '%s';`, databaseName, databaseName)

	_, err := db.ExecContext(ctx, transactionSQL)
	if err != nil {
		return fmt.Errorf("failed to terminate active connections: %w", err)
	}

	// once the connections have been removed, then we can drop the database
	_, err = db.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE %s;`, databaseName))
	if err != nil {
		return fmt.Errorf("failed to drop database %q: %w", databaseName, err)
	}

	return nil
}

// Ping calls db.Ping on the connection handle to test the connection, a simple
// wrapper.
func Ping(ctx context.Context, db *sql.DB) error {
	err := db.PingContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to ping database")
	}

	return nil
}

// Exists will return true if a database with the supplied name exists on the
// currently connected postgres instance.
func Exists(ctx context.Context, db *sql.DB, databaseName string) (bool, error) {
	// https://dba.stackexchange.com/questions/45143/check-if-postgresql-database-exists-case-insensitive-way
	rows, err := db.QueryContext(ctx, `SELECT 1 FROM pg_database WHERE datname=$1`, databaseName)
	if err != nil {
		return false, fmt.Errorf("failed look up database: %w", err)
	}
	defer rows.Close()

	var result int
	for rows.Next() {
		err = rows.Scan(&result)
		if err != nil {
			return false, fmt.Errorf(
				"failed to parse sql row response containing matching databases: %w",
				err,
			)
		}
	}

	// only return true if there was a matching row with 1 set for a name match
	return result == 1, nil
}

// buildConnectionString constructs a PostgreSQL connection string supporting both
// TCP connections (postgres://host:port/db) and Unix socket connections (postgres:///db?host=/path)
func buildConnectionString(connectionStringBase string, params url.Values) string {
	// Check if this is a Unix socket connection by looking for host parameter that starts with "/"
	if hostParam := params.Get("host"); hostParam != "" && strings.HasPrefix(hostParam, "/") {
		// Validate that the socket directory exists
		if _, err := os.Stat(hostParam); os.IsNotExist(err) {
			// Log warning but continue - PostgreSQL might create the socket
			fmt.Printf("Warning: Unix socket directory %s does not exist\n", hostParam)
		}

		// Unix socket connection: postgres:///dbname?host=/socket/path&other=params
		dbname := params.Get("dbname")
		if dbname == "" {
			dbname = "postgres"
		}
		return fmt.Sprintf("postgres:///%s?%s", dbname, params.Encode())
	}

	// Standard TCP connection: postgres://host:port/db?params
	return fmt.Sprintf("%s?%s", connectionStringBase, params.Encode())
}

// Truncate the table with tableName.
func Truncate(ctx context.Context, db *sql.DB, tableName string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf(`TRUNCATE %s CASCADE;`, tableName))
	if err != nil {
		return fmt.Errorf("failed to truncate table: %w", err)
	}

	return nil
}
