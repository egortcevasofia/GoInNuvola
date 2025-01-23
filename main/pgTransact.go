package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type PostgresTransactionLogger struct {
	events chan<- Event
	errors <-chan error
	db     *sql.DB
}

type PostgresDBParams struct {
	dbName   string
	host     string
	port     int
	user     string
	password string
}

func (l *PostgresTransactionLogger) WritePut(key, value string) {
	l.events <- Event{EventType: EventPut, Key: key, Value: value}
}

func (l *PostgresTransactionLogger) WriteDelete(key string) {
	l.events <- Event{EventType: EventDelete, Key: key}
}

func (l *PostgresTransactionLogger) Err() <-chan error {
	return l.errors
}

func NewPostgresTransactionalLogger(config PostgresDBParams) (TransactionLogger, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.host, config.port, config.user, config.password, config.dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	logger := &PostgresTransactionLogger{db: db}

	exists, err := logger.verifyTableExists()

	if err != nil {
		return nil, fmt.Errorf("failed to verify table exists: %w", err)
	}

	if !exists {
		if err = logger.createTable(); err != nil {
			return nil, fmt.Errorf("failed to create table: %w ", err)
		}
	}
	return logger, nil
}
func (l *PostgresTransactionLogger) createTable() error {
	_, err := l.db.Exec("CREATE TABLE transactions (sequence numeric, event_type varchar(100),    key varchar(100),    value varchar(100));")
	return err
}

func (l *PostgresTransactionLogger) verifyTableExists() (bool, error) {
	_, err := l.db.Query("select 1 from transactions")
	if err != nil {
		return false, nil
	} else {
		return true, nil
	}
}

func (l *PostgresTransactionLogger) Run() {
	events := make(chan Event, 16)
	l.events = events

	errors := make(chan error, 1)
	l.errors = errors
	go func() {
		query := "insert into transactions (sequence, event_type, key, value) values (nextval('seq'), $1, $2, $3)"
		for e := range events {
			_, err := l.db.Exec(query, e.EventType, e.Key, e.Value)
			if err != nil {
				errors <- err
			}
		}
	}()
}

func (l *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		defer close(outEvent)
		defer close(outError)

		query := "select sequence, event_type, key, value from transactions order by sequence"
		rows, err := l.db.Query(query)
		if err != nil {
			outError <- fmt.Errorf("sql query error: %w", err)
			return
		}
		defer rows.Close()
		e := Event{}

		for rows.Next() {
			err = rows.Scan(&e.Sequence, &e.EventType, &e.Key, &e.Value)
			if err != nil {
				outError <- fmt.Errorf("error reading row: %w", err)
				return
			}
			outEvent <- e
		}
		err = rows.Err()
		if err != nil {
			outError <- fmt.Errorf("transaction log read file error: %w", err)
		}
	}()
	return outEvent, outError
}
