package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/SubhamMurarka/KVStore/models"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type RepoInterface interface {
	Get(key string) (*models.Request, error)
	Put(inputObj *models.Request) error
	Delete(key string) error
}

type repo struct {
	connection *pgx.Conn
}

func NewRepo(conn *pgx.Conn) RepoInterface {
	return &repo{
		connection: conn,
	}
}

func (r *repo) Put(inpObj *models.Request) error {
	key := inpObj.Key
	value := inpObj.Value
	expireAt := time.Now().Add(time.Duration(inpObj.TTL) * time.Second).UTC()
	// expireAt = expireAt.Format("2006-01-02 15:04:05")
	fmt.Println("expiry : ", expireAt)
	query := `INSERT INTO kv_store(key, value, expire_at)
               VALUES($1, $2, $3)

			   `
	start := time.Now()

	op, err := r.connection.Exec(context.Background(), query, key, value, expireAt)

	duration := time.Since(start)

	if err != nil {
		logrus.Error("Error inserting/updating: ", err)
		return err
	}

	fmt.Println("time taken : ", duration)

	logrus.Infof("Rows affected by Put operation: %d", op.RowsAffected())
	return nil
}

// func (r *repo) Put(inpObj *models.Request) error {
// 	key := inpObj.Key
// 	value := inpObj.Value
// 	expireAt := time.Now().Add(time.Duration(inpObj.TTL) * time.Second)

// 	query := `UPDATE kv_store
//               SET value = COALESCE($2, kv_store.value),
//                   expire_at = COALESCE($3, kv_store.expire_at)
//               WHERE key = $1 and expire_at > NOW();`

// 	op, err := r.connection.Exec(context.Background(), query, key, value, expireAt)
// 	if err != nil {
// 		logrus.Error("Error updating: ", err)
// 		return err
// 	}

// 	// Check if any rows were affected
// 	if op.RowsAffected() == 0 {
// 		return fmt.Errorf("no record found with key: %s", key)
// 	}

// 	logrus.Infof("Rows affected by Update operation: %d", op.RowsAffected())
// 	return nil
// }

func (r *repo) Get(key string) (*models.Request, error) {
	kvop := &models.Request{}

	query := `SELECT key, value, expire_at FROM kv_store
              WHERE key = $1 and expire_at > NOW()`

	row := r.connection.QueryRow(context.Background(), query, key)

	var expireAt time.Time

	start := time.Now()

	err := row.Scan(&kvop.Key, &kvop.Value, &expireAt)
	duration := time.Since(start)
	if err != nil {
		if err == pgx.ErrNoRows {
			logrus.Info("No rows found for key: ", key)
			return nil, nil
		}
		logrus.Error("Error retrieving key: ", key, err)
		return nil, err
	}

	kvop.TTL = int64(expireAt.Sub(time.Now()).Seconds())

	fmt.Println("time taken : ", duration)

	logrus.Info("Row found for key: ", key)
	return kvop, nil
}

func (r *repo) Delete(key string) error {
	query := `UPDATE kv_store
		SET expire_at = '4713-11-24 00:00:00 BC'
		WHERE key = $1 AND expire_at > NOW();`

	start := time.Now()

	op, err := r.connection.Exec(context.Background(), query, key)

	duration := time.Since(start)

	if err != nil {
		logrus.Error("Error in deleting: ", err)
		return err
	}

	fmt.Println("time taken : ", duration)

	logrus.Infof("Rows deleted for key %s: %d", key, op.RowsAffected())
	return nil
}
