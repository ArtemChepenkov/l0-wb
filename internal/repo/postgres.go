package repo

import (
	"database/sql"
	"errors"
	_ "fmt"
	_ "io/ioutil"
	"time"

	"l0-wb/internal/model"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const migrationsPath = "migrations"

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo { return &PostgresRepo{db: db} }

func RunMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, "postgres", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func (r *PostgresRepo) SaveOrder(o model.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (order_uid) DO UPDATE SET track_number = EXCLUDED.track_number`,
		o.OrderUID, o.TrackNumber, o.Entry, o.Locale, o.InternalSignature, o.CustomerID, o.DeliveryService, o.ShardKey, o.SmID, o.DateCreated, o.OofShard)
	if err != nil {
		return err
	}

	if o.Delivery != nil {
		_, err = tx.Exec(`INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT (order_uid) DO UPDATE SET name = EXCLUDED.name`,
			o.OrderUID, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip, o.Delivery.City, o.Delivery.Address, o.Delivery.Region, o.Delivery.Email)
		if err != nil {
			return err
		}
	}

	if o.Payment != nil {
		_, err = tx.Exec(`INSERT INTO payment (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT (order_uid) DO UPDATE SET transaction = EXCLUDED.transaction`,
			o.OrderUID, o.Payment.Transaction, o.Payment.RequestID, o.Payment.Currency, o.Payment.Provider, o.Payment.Amount, o.Payment.PaymentDT, o.Payment.Bank, o.Payment.DeliveryCost, o.Payment.GoodsTotal, o.Payment.CustomFee)
		if err != nil {
			return err
		}
	}

	for _, it := range o.Items {
		_, err = tx.Exec(`INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			o.OrderUID, it.ChrtID, it.TrackNumber, it.Price, it.Rid, it.Name, it.Sale, it.Size, it.TotalPrice, it.NmID, it.Brand, it.Status)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepo) GetLatest(limit int) ([]model.Order, error) {
	rows, err := r.db.Query(`SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders ORDER BY date_created DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []model.Order
	for rows.Next() {
		var o model.Order
		var dt time.Time
		if err := rows.Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature, &o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SmID, &dt, &o.OofShard); err != nil {
			return nil, err
		}
		o.DateCreated = dt
		res = append(res, o)
	}
	return res, nil
}

func (r *PostgresRepo) GetFull(orderUID string) (model.Order, error) {
	var o model.Order
	var dt time.Time
	err := r.db.QueryRow(`SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE order_uid=$1`, orderUID).
		Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature, &o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SmID, &dt, &o.OofShard)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Order{}, sql.ErrNoRows
		}
		return model.Order{}, err
	}
	o.DateCreated = dt

	var d model.Delivery
	if err := r.db.QueryRow(`SELECT name, phone, zip, city, address, region, email FROM delivery WHERE order_uid=$1`, orderUID).
		Scan(&d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email); err == nil {
		o.Delivery = &d
	}

	var p model.Payment
	if err := r.db.QueryRow(`SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payment WHERE order_uid=$1`, orderUID).
		Scan(&p.Transaction, &p.RequestID, &p.Currency, &p.Provider, &p.Amount, &p.PaymentDT, &p.Bank, &p.DeliveryCost, &p.GoodsTotal, &p.CustomFee); err == nil {
		o.Payment = &p
	}

	rows, err := r.db.Query(`SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid=$1`, orderUID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var it model.Item
			if err := rows.Scan(&it.ChrtID, &it.TrackNumber, &it.Price, &it.Rid, &it.Name, &it.Sale, &it.Size, &it.TotalPrice, &it.NmID, &it.Brand, &it.Status); err == nil {
				o.Items = append(o.Items, it)
			}
		}
	}

	return o, nil
}
