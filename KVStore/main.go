package main

import (
	"context"
	"fmt"

	"github.com/SubhamMurarka/KVStore/handler"
	"github.com/SubhamMurarka/KVStore/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	DBURL := "postgres://kv_admin:kv_password@localhost:5432/KV"

	conn, err := pgx.Connect(context.Background(), DBURL)
	if err != nil {
		logrus.Fatal("DB connection refused", err)
	}

	fmt.Println("DB connected")

	defer conn.Close(context.Background())

	repo := repository.NewRepo(conn)
	handle := handler.NewHandler(repo)

	r := gin.Default()
	r.POST("/put", handle.Put)
	r.GET("/get", handle.Get)
	r.DELETE("/delete", handle.Delete)

	r.Run("0.0.0.0:8080")
}
