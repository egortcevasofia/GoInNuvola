package transact

import (
	"GoInNuvola/core"
	"fmt"
	"os"
)

func NewTransactionLogger(logger string) (core.TransactionLogger, error) {
	switch logger {
	case "file":
		return NewFileTransactionalLogger(os.Getenv("TLOG_FILENAME"))
	case "postgres":
		return NewPostgresTransactionalLogger(PostgresDBParams{
			DbName:   "postgres",
			Host:     "172.17.0.1",
			Port:     5432,
			User:     "postgres",
			Password: "admin",
		})
	case "":
		return nil, fmt.Errorf("transaction logger type not defined")
	default:
		return nil, fmt.Errorf("transaction logger type not defined")

	}
}
