package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Fserlut/gophermart/internal/lib"
	_ "github.com/lib/pq"
	"log"

	"github.com/Fserlut/gophermart/internal/models"
)

type Database struct {
	db *sql.DB
}

func (d *Database) CreateUser(userCreate models.User) error {
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

func (d *Database) GetOrderByNumber(orderNumber string) (*models.Order, error) {
	return nil, nil
}

func (d *Database) CreateOrder(orderNumber string, userUUID string) error {
	res, err := d.db.ExecContext(
		context.Background(),
		`INSERT INTO orders (number, user_uuid) VALUES ($1, $2) ON CONFLICT (number) DO NOTHING`, orderNumber, userUUID,
	)

	if err != nil {
		return err
	}

	affectedRows, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if affectedRows == 0 {
		var userID string
		row := d.db.QueryRowContext(
			context.Background(),
			"SELECT user_uuid FROM orders WHERE number = $1", orderNumber,
		)
		err := row.Scan(&userID)

		if err != nil {
			return err
		}
		if userID != userUUID {
			return &lib.ErrOrderAlreadyCreatedByOtherUser{}
		}
		return &lib.ErrOrderAlreadyCreated{}
	}

	return nil
}

func (d *Database) GetOrdersByUserID(userID string) ([]models.Order, error) {
	query := `SELECT number, status, uploaded_at, accrual FROM orders WHERE user_uuid = $1`

	rows, err := d.db.QueryContext(context.Background(), query, userID)
	if err != nil {
		log.Printf("Error querying orders for userID %s: %v", userID, err)
		return nil, fmt.Errorf("error querying orders for userID %s: %w", userID, err)
	}
	defer rows.Close()

	var orders []models.Order

	for rows.Next() {
		var order models.Order
		var accrual sql.NullInt64
		if err := rows.Scan(&order.Number, &order.Status, &order.UploadedAt, &accrual); err != nil {
			log.Printf("Error scanning order for userID %s: %v", userID, err)
			return nil, err
		}
		if accrual.Valid {
			order.Accrual = &accrual.Int64
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating orders for user_ID %s: %v", userID, err)
		return nil, fmt.Errorf("error iterating orders for userID %s: %w", userID, err)
	}

	return orders, nil
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

		CREATE TABLE IF NOT EXISTS orders (
				number TEXT PRIMARY KEY,
				user_uuid TEXT,
				status TEXT DEFAULT 'NEW',
				accrual INTEGER,
				uploaded_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_uuid) REFERENCES users(uuid)
		);
	`)

	if err != nil {
		return nil, err
	}

	return &Database{
		db: db,
	}, nil
}
