package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	DBURL := "postgresql://user:password@localhost:5432/postgres"

	pgconn, err := pgx.Connect(context.Background(), DBURL)
	if err != nil {
		logrus.Fatal("DB connection refused", err)
	}

	logrus.Info("DB connected")

	defer pgconn.Close(context.Background())

	ticker := time.NewTicker(5 * time.Second)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cleanUp(pgconn)
		}
	}
}

func cleanUp(pgconn *pgx.Conn) {
	query := `DELETE FROM kv_store WHERE expire_at <= NOW();`

	pgcmd, err := pgconn.Exec(context.Background(), query)
	if err != nil {
		logrus.Error("Error Deleting : ", err)
	}

	logrus.Info("No. of rows affected : ", pgcmd.RowsAffected())
}
