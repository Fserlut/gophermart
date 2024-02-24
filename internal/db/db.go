package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Fserlut/gophermart/internal/lib"
	_ "github.com/lib/pq"

	"github.com/Fserlut/gophermart/internal/models"
)

type Database struct {
	db *sql.DB
}

func (d *Database) Create(userCreate models.User) error {
	res, err := d.db.ExecContext(
		context.Background(),
		`INSERT INTO users (uuid, login, password) VALUES ($1, $2, $3) ON CONFLICT (login) DO NOTHING`, userCreate.UUID, userCreate.Login, userCreate.Password,
	)

	if err != nil {
		return err
	}

	affectedRows, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if affectedRows == 0 {
		return &lib.ErrUserExists{}
	}

	return nil
}

func (d *Database) GetUserByLogin(loginToFind string) (*models.User, error) {
	var (
		uuid     string
		login    string
		password string
	)
	row := d.db.QueryRowContext(
		context.Background(),
		"SELECT uuid, login, password FROM users WHERE login = $1", loginToFind,
	)
	err := row.Scan(&uuid, &login, &password)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &lib.ErrUserNotFound{}
		} else {
			return nil, err
		}
	}

	return &models.User{
		UUID:     uuid,
		Login:    login,
		Password: password,
	}, nil
}

func NewDB(dsn string) (*Database, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
        uuid TEXT PRIMARY KEY,
        login TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL
    );

		CREATE UNIQUE INDEX IF NOT EXISTS user_uniq_index
		    on users (login);
	`)

	if err != nil {
		return nil, err
	}

	return &Database{
		db: db,
	}, nil
}
