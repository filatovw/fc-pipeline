package config

import "fmt"

type Queue struct {
	Addr string
	User string
	Pass string
}

func (q Queue) ConnectionString() string {
	return fmt.Sprintf("amqp://%s:%s@%s/", q.User, q.Pass, q.Addr)
}

type DB struct {
	Host string
	Port int
	User string
	Pass string
}

func (db DB) ConnectionString(dbname string) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		db.Host, db.Port, db.User, db.Pass, dbname)
}
