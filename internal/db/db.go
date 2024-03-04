package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/Fserlut/gophermart/internal/lib"
	"github.com/Fserlut/gophermart/internal/models/order"
	"github.com/Fserlut/gophermart/internal/models/user"
)

type Database struct {
	db *sql.DB
}

func (d *Database) CreateUser(userCreate user.User) error {
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

func (d *Database) GetUserByLogin(loginToFind string) (*user.User, error) {
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

	return &user.User{
		UUID:     uuid,
		Login:    login,
		Password: password,
	}, nil
}

func (d *Database) GetOrderByNumber(orderNumber string) (*order.Order, error) {
	return nil, nil
}

func (d *Database) CreateOrder(orderNumber string, UserUUID string, withdraw *float64) error {
	res, err := d.db.ExecContext(
		context.Background(),
		`INSERT INTO orders (number, user_uuid, withdraw) VALUES ($1, $2, $3) ON CONFLICT (number) DO NOTHING`, orderNumber, UserUUID, withdraw,
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
		if userID != UserUUID {
			return &lib.ErrOrderAlreadyCreatedByOtherUser{}
		}
		return &lib.ErrOrderAlreadyCreated{}
	}

	return nil
}

func (d *Database) GetOrdersByUserID(userID string) ([]order.Order, error) {
	query := `SELECT number, status, uploaded_at, accrual FROM orders WHERE user_uuid = $1`

	rows, err := d.db.QueryContext(context.Background(), query, userID)
	if err != nil {
		log.Printf("Error querying orders for userID %s: %v", userID, err)
		return nil, fmt.Errorf("error querying orders for userID %s: %w", userID, err)
	}
	defer rows.Close()

	var orders []order.Order

	for rows.Next() {
		var order order.Order
		var accrual sql.NullFloat64
		if err := rows.Scan(&order.Number, &order.Status, &order.UploadedAt, &accrual); err != nil {
			log.Printf("Error scanning order for userID %s: %v", userID, err)
			return nil, err
		}
		if accrual.Valid {
			order.Accrual = &accrual.Float64
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating orders for userID %s: %v", userID, err)
		return nil, fmt.Errorf("error iterating orders for userID %s: %w", userID, err)
	}

	return orders, nil
}

func (d *Database) GetUserBalance(userID string) (*user.BalanceResponse, error) {
	var response user.BalanceResponse

	query := `
	SELECT
		COALESCE(SUM(accrual)::NUMERIC(10,2), 0) AS current,
		COALESCE(SUM(withdraw)::NUMERIC(10,2), 0) AS withdrawn
	FROM orders
	WHERE user_uuid = $1
	`

	err := d.db.QueryRow(query, userID).Scan(&response.Current, &response.Withdrawn)
	if err != nil {
		return nil, fmt.Errorf("error querying user balance: %w", err)
	}

	return &user.BalanceResponse{
		Current:   response.Current - response.Withdrawn,
		Withdrawn: response.Withdrawn,
	}, nil
}

func (d *Database) Withdrawals(userID string) ([]user.WithdrawalsResponse, error) {
	query := `SELECT number, withdraw, processed_at FROM orders WHERE user_uuid = $1 AND withdraw > 0`

	rows, err := d.db.QueryContext(context.Background(), query, userID)
	if err != nil {
		log.Printf("Error querying withdrawals for userID %s: %v", userID, err)
		return nil, fmt.Errorf("error querying withdrawals for userID %s: %w", userID, err)
	}
	defer rows.Close()

	var res []user.WithdrawalsResponse

	for rows.Next() {
		var resItem user.WithdrawalsResponse
		if err := rows.Scan(&resItem.Order, &resItem.Sum, &resItem.ProcessedAt); err != nil {
			log.Printf("Error scanning order for userID %s: %v", userID, err)
			return nil, err
		}
		res = append(res, resItem)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating orders for userID %s: %v", userID, err)
		return nil, fmt.Errorf("error iterating orders for userID %s: %w", userID, err)
	}

	return res, nil
}

func (d *Database) Update(orderNumber string, status string, accrual *float64) error {
	var query string
	if accrual != nil {
		query = `UPDATE orders SET status = $2, accrual = $3 WHERE number = $1`
	} else {
		query = `UPDATE orders SET status = $2 WHERE number = $1`
	}

	var err error
	if accrual != nil {
		_, err = d.db.Exec(query, orderNumber, status, *accrual)
	} else {
		_, err = d.db.Exec(query, orderNumber, status)
	}

	return err
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
        ON users (login);

    CREATE TABLE IF NOT EXISTS orders (
        number TEXT PRIMARY KEY,
        user_uuid TEXT,
        status TEXT DEFAULT 'NEW',
        accrual NUMERIC(10, 2),
        withdraw NUMERIC(10, 2),
        uploaded_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        processed_at TIMESTAMP,
        FOREIGN KEY (user_uuid) REFERENCES users(uuid)
    );

    CREATE OR REPLACE FUNCTION update_processed_at()
    RETURNS TRIGGER AS $$
    BEGIN
        IF NEW.withdraw IS NOT NULL AND OLD.withdraw IS DISTINCT FROM NEW.withdraw THEN
            NEW.processed_at = CURRENT_TIMESTAMP;
        ELSIF TG_OP = 'INSERT' AND NEW.withdraw IS NOT NULL THEN
            NEW.processed_at = CURRENT_TIMESTAMP;
        END IF;
        RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;

    DROP TRIGGER IF EXISTS trigger_update_processed_at ON orders;
    CREATE TRIGGER trigger_update_processed_at
    BEFORE INSERT OR UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION update_processed_at();
`)

	if err != nil {
		return nil, err
	}

	return &Database{
		db: db,
	}, nil
}
