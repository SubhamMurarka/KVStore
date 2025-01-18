package config

import "github.com/jackc/pgx/v5/pgxpool"

var numshards = 2

var writePool[numshards] *pgxpool.Pool
var writePool[numshards] *pgxpool.Pool

func connect() {
	DBWriteURL_1 := "postgresql://user:password@localhost:5432/postgres"
	DBReadURL_1 := "postgresql://replicator:replicator_password@localhost:5433/postgres" // Replica DB

	DBWriteURL_2 := "postgresql://user:password@localhost:5434/postgres"
	DBReadURL_2 := "postgresql://replicator:replicator_password@localhost:5435/postgres"

	for i := 0; i < numshards; i++ {
		writeConfig, err := pgxpool.ParseConfig(DBWriteURL)
		if err != nil {
			logrus.Fatalf("Unable to configure primary database: %v", err)
		}

		writeConfig.MinConns = 5
		writeConfig.MaxConns = 60
	}

}

// Configure primary (write) connection pool
writeConfig, err := pgxpool.ParseConfig(DBWriteURL)
if err != nil {
	logrus.Fatalf("Unable to configure primary database: %v", err)
}

writeConfig.MinConns = 5
writeConfig.MaxConns = 60

writePool, err := pgxpool.NewWithConfig(context.Background(), writeConfig)
if err != nil {
	logrus.Fatalf("Unable to create write connection pool: %v", err)
}
defer writePool.Close()

// Configure replica (read) connection pool
readConfig, err := pgxpool.ParseConfig(DBReadURL)
if err != nil {
	logrus.Fatalf("Unable to configure replica database: %v", err)
}

readConfig.MinConns = 5
readConfig.MaxConns = 60

readPool, err := pgxpool.NewWithConfig(context.Background(), readConfig)
if err != nil {
	logrus.Fatalf("Unable to create read connection pool: %v", err)
}
defer readPool.Close()