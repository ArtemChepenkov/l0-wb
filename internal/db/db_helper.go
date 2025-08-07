package db

import (
	"database/sql"
	"log"
    "fmt"
)

func GetFullInfo(db *sql.DB, orderUID string) (*Order, error) {

	order := Order{}
	err := db.QueryRow(`SELECT order_uid, track_number, entry, locale, internal_signature,
        customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders WHERE order_uid = $1`, orderUID).
        Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
            &order.InternalSignature, &order.CustomerID, &order.DeliveryService,
            &order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard)

    if err != nil {
		log.Printf("Failed to get orders: %v\n", err)
        return nil, err
    }

	delivery := Delivery{}
	err = db.QueryRow(`SELECT name, phone, zip, city, address, region, email
        FROM delivery WHERE order_uid = $1`, orderUID).
        Scan(&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City,
            &delivery.Address, &delivery.Region, &delivery.Email)

    if err == nil {
        order.Delivery = &delivery
    } else {
		log.Printf("Failed to get delivery: %v\n", err)
	}

	payment := Payment{}
    err = db.QueryRow(`SELECT transaction, request_id, currency, provider, amount,
        payment_dt, bank, delivery_cost, goods_total, custom_fee
        FROM payment WHERE order_uid = $1`, orderUID).
        Scan(&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider,
            &payment.Amount, &payment.PaymentDT, &payment.Bank, &payment.DeliveryCost,
            &payment.GoodsTotal, &payment.CustomFee)
    if err == nil {
        order.Payment = &payment
    } else {
		log.Printf("Failed to get payment: %v\n", err)
	}

	rows, err := db.Query(`SELECT chrt_id, track_number, price, rid, name, sale,
        size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1`, orderUID)
    if err == nil {
        defer rows.Close()
        for rows.Next() {
            var item Item
            err = rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
            &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID,
            &item.Brand, &item.Status)
            if err == nil {
                order.Items = append(order.Items, item)
            } else {
				
			}
        }
	} else {
		log.Printf("Failed to get items: %v\n", err)
	}
	return &order, nil
}

func LoadFullINfo(db *sql.DB, order Order) error {
    tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w\n", err)
	}

	_, err = tx.Exec(`INSERT INTO orders (
		order_uid, track_number, entry, locale, internal_signature,
		customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("insert into orders: %w\n", err)
	}


	if order.Delivery != nil {
		_, err = tx.Exec(`INSERT INTO delivery (
			order_uid, name, phone, zip, city, address, region, email
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
			order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("insert into delivery: %w\n", err)
		}
	}

	if order.Payment != nil {
		_, err = tx.Exec(`INSERT INTO payment (
			order_uid, transaction, request_id, currency, provider, amount,
			payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
			order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
			order.Payment.PaymentDT, order.Payment.Bank, order.Payment.DeliveryCost,
			order.Payment.GoodsTotal, order.Payment.CustomFee)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("insert into payment: %w\n", err)
		}
	}

	for _, item := range order.Items {
		_, err = tx.Exec(`INSERT INTO items (
			order_uid, chrt_id, track_number, price, rid, name, sale,
			size, total_price, nm_id, brand, status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid,
			item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("insert into items: %w\n", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w\n", err)
	}

	return nil
}