package main

import (
	"context"

	"github.com/SubhamMurarka/KVStore/handler"
	"github.com/SubhamMurarka/KVStore/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	DBWriteURL := "postgresql://user:password@localhost:5432/postgres"
	DBReadURL := "postgresql://replicator:replicator_password@localhost:5433/postgres" // Replica DB

	// Configure primary (write) connection pool
	writeConfig, err := pgxpool.ParseConfig(DBWriteURL)
	if err != nil {
		logrus.Fatalf("Unable to configure primary database: %v", err)
	}

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

	readPool, err := pgxpool.NewWithConfig(context.Background(), readConfig)
	if err != nil {
		logrus.Fatalf("Unable to create read connection pool: %v", err)
	}
	defer readPool.Close()

	repo := repository.NewRepo(writePool, readPool)
	handle := handler.NewHandler(repo)

	r := gin.Default()
	r.POST("/put", handle.Put)
	r.GET("/get", handle.Get)
	r.DELETE("/delete", handle.Delete)
	r.PATCH("/update", handle.Update)
	r.GET("/pool-stats", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"write_pool": gin.H{
				"total_conns":    writePool.Stat().TotalConns(),
				"idle_conns":     writePool.Stat().IdleConns(),
				"acquired_conns": writePool.Stat().AcquiredConns(),
				"max_conns":      writePool.Stat().MaxConns(),
			},
			"read_pool": gin.H{
				"total_conns":    readPool.Stat().TotalConns(),
				"idle_conns":     readPool.Stat().IdleConns(),
				"acquired_conns": readPool.Stat().AcquiredConns(),
				"max_conns":      readPool.Stat().MaxConns(),
			},
		})
	})

	r.Run("0.0.0.0:8080")
}
