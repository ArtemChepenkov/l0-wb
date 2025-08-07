-- Сначала orders
INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature,
    customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
VALUES 
('b563feb7b2b84b6test1', 'WBILMTESTTRACK', 'WBIL', 'en', '',
 'test123', 'meest', '9', 99, '2021-11-26T06:22:19Z', '1');

-- Потом delivery
INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
VALUES 
('b563feb7b2b84b6test1', 'Иван Иванов', '+79001234567', '123456', 'Москва', 'ул. Пушкина, д.1', 'Московская область', 'ivan@example.com');

-- Потом payment
INSERT INTO payment (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
VALUES 
('b563feb7b2b84b6test1', 'b563feb7b2b84b6test1', '', 'RUB', 'yoomoney', 1500, 1637907727, 'Sberbank', 200, 1300, 0);

-- Потом items
INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size,
    total_price, nm_id, brand, status)
VALUES 
('b563feb7b2b84b6test1', 9934930, 'WBILMTESTTRACK', 1500, 'ab4219087a764ae0btest', 'Маска', 0, '0',
 1500, 2389212, 'MaskBrand', 202);

