package main

import (
	"encoding/binary"
	"encoding/json"
	"time"

	bolt "go.etcd.io/bbolt"
)

const (
	dbName     = "tasks.db"
	bucketName = "Tasks"
)

type Task struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	PickedAt    time.Time `json:"picked_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

func (t Task) Duration() time.Duration {
	if t.CompletedAt.IsZero() || t.PickedAt.IsZero() {
		return 0
	}
	return t.CompletedAt.Sub(t.PickedAt)
}

// DB Helpers

func setupDB() (*bolt.DB, error) {
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	return db, err
}

func addTask(db *bolt.DB, name string) (Task, error) {
	var task Task
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		id64, _ := b.NextSequence()
		id := int(id64)
		
		task = Task{
			ID:   id,
			Name: name,
		}
		
		buf, err := json.Marshal(task)
		if err != nil {
			return err
		}
		
		return b.Put(itob(id), buf)
	})
	return task, err
}

func updateTask(db *bolt.DB, task Task) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		
		buf, err := json.Marshal(task)
		if err != nil {
			return err
		}
		
		return b.Put(itob(task.ID), buf)
	})
}

func loadTasks(db *bolt.DB) ([]Task, error) {
	var tasks []Task
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var t Task
			if err := json.Unmarshal(v, &t); err != nil {
				// Skip invalid entries or handle error
				continue 
			}
			tasks = append(tasks, t)
		}
		return nil
	})
	return tasks, err
}

func deleteTask(db *bolt.DB, id int) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.Delete(itob(id))
	})
}

func clearTasks(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte(bucketName)); err != nil {
			return err
		}
		_, err := tx.CreateBucket([]byte(bucketName))
		return err
	})
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
