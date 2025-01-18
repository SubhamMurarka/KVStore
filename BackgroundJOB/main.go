package main

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

var wg sync.WaitGroup

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	DBURL1 := "postgresql://user:password@localhost:5432/postgres"
	DBURL2 := "postgresql://user:password@localhost:5434/postgres"

	pgconn, err := pgx.Connect(context.Background(), DBURL1)
	if err != nil {
		logrus.Fatal("DB connection refused", err)
	}

	logrus.Info("DB connected")

	defer pgconn.Close(context.Background())

	pgconn2, err := pgx.Connect(context.Background(), DBURL2)
	if err != nil {
		logrus.Fatal("DB connection refused", err)
	}

	logrus.Info("DB connected")

	defer pgconn2.Close(context.Background())

	ticker := time.NewTicker(5 * time.Second)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wg.Add(2)
			go func() {
				defer wg.Done()
				cleanUp(pgconn)
			}()
			go func() {
				defer wg.Done()
				cleanUp(pgconn2)
			}()
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
