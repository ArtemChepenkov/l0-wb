INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
VALUES ('abc123', 'TRACK123', 'web', 'en', '', 'cust1', 'dhl', 'key1', 123, NOW(), 'shard1');

INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
VALUES ('abc123', 'John Doe', '1234567890', '123456', 'New York', '5th Ave 10', 'NY', 'john@example.com');

INSERT INTO payment (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
VALUES ('abc123', 'trans123', 'req123', 'USD', 'stripe', 1500, 1620000000, 'bank1', 100, 1400, 0);

INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
VALUES 
('abc123', 1, 'TRACK123', 500, 'RID123', 'T-Shirt', 10, 'M', 450, 1001, 'Nike', 1),
('abc123', 2, 'TRACK123', 1000, 'RID124', 'Shoes', 5, '42', 950, 1002, 'Adidas', 1);
