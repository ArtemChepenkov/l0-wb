-- delivery
INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
VALUES 
('b563feb7b2b84b6test', 'Иван Иванов', '+79001234567', '123456', 'Москва', 'ул. Пушкина, д.1', 'Московская область', 'ivan@example.com');

-- payment
INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee, order_uid)
VALUES 
('b563feb7b2b84b6test', '', 'RUB', 'yoomoney', 1500, 1637907727, 'Sberbank', 200, 1300, 0, 'b563feb7b2b84b6test');

-- orders
INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature,
    customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
VALUES 
('b563feb7b2b84b6test', 'WBILMTESTTRACK', 'WBIL', 'en', '',
 'test123', 'meest', '9', 99, '2021-11-26T06:22:19Z', '1');

-- items
INSERT INTO items (chrt_id, track_number, price, rid, name, sale, size,
    total_price, nm_id, brand, status, order_uid)
VALUES 
(9934930, 'WBILMTESTTRACK', 1500, 'ab4219087a764ae0btest', 'Маска', 0, '0',
 1500, 2389212, 'MaskBrand', 202, 'b563feb7b2b84b6test');
