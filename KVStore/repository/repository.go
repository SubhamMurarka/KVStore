package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/SubhamMurarka/KVStore/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/spaolacci/murmur3"
)

type RepoInterface interface {
	Get(key string) (*models.Request, error)
	Put(inputObj *models.Request) error
	Delete(key string) error
	Update(inputObj *models.UpdateRequest) error
}

type repo struct {
	connectionWrite []*pgxpool.Pool
	connectionRead  []*pgxpool.Pool
}

func NewRepo(connWrite []*pgxpool.Pool, connRead []*pgxpool.Pool) RepoInterface {
	return &repo{
		connectionWrite: connWrite,
		connectionRead:  connRead,
	}
}

func GetShardID(key string) uint32 {
	hash := murmur3.Sum32([]byte(key))

	// Shard the data based on the hash value
	numShards := 2
	shardID := hash % uint32(numShards)

	fmt.Printf("Hash: %d, Shard: %d\n", hash, shardID)

	return shardID
}

func (r *repo) Put(inpObj *models.Request) error {
	key := inpObj.Key
	i := GetShardID(key)
	value := inpObj.Value
	expireAt := time.Now().Add(time.Duration(inpObj.TTL) * time.Second).UTC()

	fmt.Println("expiry : ", expireAt)
	query := `INSERT INTO kv_store(key, value, expire_at)
               VALUES($1, $2, $3)

			   `
	start := time.Now()

	op, err := r.connectionWrite[i].Exec(context.Background(), query, key, value, expireAt)

	duration := time.Since(start)

	if err != nil {
		logrus.Error("Error inserting/updating: ", err)
		return err
	}

	fmt.Println("time taken : ", duration)

	logrus.Infof("Rows affected by Put operation: %d", op.RowsAffected())
	return nil
}

func (r *repo) Get(key string) (*models.Request, error) {
	i := GetShardID(key)
	kvop := &models.Request{}

	query := `SELECT key, value, expire_at FROM kv_store
              WHERE key = $1 and expire_at > NOW()`

	row := r.connectionRead[i].QueryRow(context.Background(), query, key)

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
	i := GetShardID(key)

	query := `UPDATE kv_store
		SET expire_at = '4713-11-24 00:00:00 BC'
		WHERE key = $1 AND expire_at > NOW();`

	start := time.Now()

	op, err := r.connectionWrite[i].Exec(context.Background(), query, key)

	duration := time.Since(start)

	if err != nil {
		logrus.Error("Error in deleting: ", err)
		return err
	}

	fmt.Println("time taken : ", duration)

	logrus.Infof("Rows deleted for key %s: %d", key, op.RowsAffected())
	return nil
}

func (r *repo) Update(inputObj *models.UpdateRequest) error {
	i := GetShardID(inputObj.Key)

	start := time.Now()
	tx, err := r.connectionWrite[i].BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		logrus.Error("not able to begin transaction : ", err)
		return err
	}

	defer tx.Rollback(context.Background())

	query := `SELECT expire_at FROM kv_store 
			  WHERE key = $1
			  FOR UPDATE
			  `
	var expire_at time.Time

	err = tx.QueryRow(context.Background(), query, inputObj.Key).Scan(&expire_at)

	if err == pgx.ErrNoRows {
		logrus.Error("key does not exist : ", inputObj.Key)
		return fmt.Errorf("key does not exist")
	}

	if err != nil {
		logrus.Error("Error querying kv_store : ", err)
		return err
	}

	if expire_at.Before(time.Now()) {
		logrus.Error("Key is expired or not eligible for update, expire_at: ", expire_at)
		return fmt.Errorf("key is expired or not eligible for update")
	}

	updateQuery := `
        UPDATE kv_store
        SET value = $1
        WHERE key = $2
    `

	pgcmd, err := tx.Exec(context.Background(), updateQuery, inputObj.Value, inputObj.Key)
	if err != nil {
		logrus.Error("Error updating kv_store: ", err)
		return err
	}

	logrus.Info("NO. of rows affected : ", pgcmd.RowsAffected())

	err = tx.Commit(context.Background())

	duration := time.Since(start)
	fmt.Println("time taken : ", duration)

	if err != nil {
		logrus.Error("Error commiting transaction : ", err)
		return err
	}

	return nil
}
