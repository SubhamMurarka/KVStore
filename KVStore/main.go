package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/SubhamMurarka/KVStore/handler"
	"github.com/SubhamMurarka/KVStore/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

var writePool []*pgxpool.Pool
var readPool []*pgxpool.Pool

func main() {
	DBShard1 := []string{
		"postgresql://user:password@postgres_primary:5432/postgres",
		"postgresql://replicator:replicator_password@postgres_replica:5432/postgres",
	}

	DBShard2 := []string{
		"postgresql://user:password@postgres_primary_1:5432/postgres",
		"postgresql://replicator:replicator_password@postgres_replica_1:5432/postgres",
	}

	// Configure primary (write) connection pool

	connect(DBShard1)
	connect(DBShard2)

	defer readPool[0].Close()
	defer readPool[1].Close()
	defer writePool[0].Close()
	defer writePool[1].Close()

	repo := repository.NewRepo(writePool, readPool)
	handle := handler.NewHandler(repo)

	r := gin.Default()
	r.POST("/put", handle.Put)
	r.GET("/get", handle.Get)
	r.DELETE("/delete", handle.Delete)
	r.PATCH("/update", handle.Update)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"response": "healthy"})
	})
	r.GET("/pool-stats:i", func(c *gin.Context) {
		i, _ := strconv.Atoi(c.Param("i"))
		c.JSON(200, gin.H{
			"write_pool": gin.H{
				"total_conns":    writePool[i].Stat().TotalConns(),
				"idle_conns":     writePool[i].Stat().IdleConns(),
				"acquired_conns": writePool[i].Stat().AcquiredConns(),
				"max_conns":      writePool[i].Stat().MaxConns(),
			},
			"read_pool": gin.H{
				"total_conns":    readPool[i].Stat().TotalConns(),
				"idle_conns":     readPool[i].Stat().IdleConns(),
				"acquired_conns": readPool[i].Stat().AcquiredConns(),
				"max_conns":      readPool[i].Stat().MaxConns(),
			},
		})
	})

	r.Run("0.0.0.0:8080")
}

func connect(DBshard []string) {
	writeConfig, err := pgxpool.ParseConfig(DBshard[0])
	if err != nil {
		logrus.Fatalf("Unable to configure primary database: %v", err)
	}

	writeConfig.MinConns = 5
	writeConfig.MaxConns = 45

	writepool, err := pgxpool.NewWithConfig(context.Background(), writeConfig)
	if err != nil {
		logrus.Fatalf("Unable to create write connection pool: %v", err)
	}
	writePool = append(writePool, writepool)

	// Configure replica (read) connection pool
	readConfig, err := pgxpool.ParseConfig(DBshard[1])
	if err != nil {
		logrus.Fatalf("Unable to configure replica database: %v", err)
	}

	readConfig.MinConns = 5
	readConfig.MaxConns = 47

	readpool, err := pgxpool.NewWithConfig(context.Background(), readConfig)
	if err != nil {
		logrus.Fatalf("Unable to create read connection pool: %v", err)
	}

	readPool = append(readPool, readpool)
}
