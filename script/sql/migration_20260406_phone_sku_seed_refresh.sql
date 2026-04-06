START TRANSACTION;

UPDATE pms_product_spu
SET name = '小米（MI）REDMI Note15 Pro 天玑7400-Ultra 7000mAh 龙晶玻璃十倍抗摔 IP68 红米 5G手机'
WHERE id = 1;

UPDATE pms_product_spu
SET name = 'Apple/苹果 iPhone 17 支持移动联通电信5G 双卡双待手机'
WHERE id = 2;

UPDATE pms_product_spu
SET name = '华为（HUAWEI）Mate X5 典藏版折叠屏手机全网通正品特北斗卫星华为大折叠'
WHERE id = 3;

UPDATE pms_product_spu
SET name = '三星（SAMSUNG）W25 心系天下 折叠屏手机 2 亿像素 Galaxy AI 商务智能手机'
WHERE id = 4;

UPDATE pms_product_spu
SET name = '荣耀X70 金标十面抗摔 8300mAh 青海湖电池 IP69 防水 AI 手机'
WHERE id = 5;

UPDATE pms_product_sku
SET name = '小米手机 银色 128GB 全网通版',
    sku_code = 'XM-128-SLV',
    price = 7999.00,
    promotion_price = 7899.00,
    stock = 1000,
    low_stock = 100,
    spec_data = JSON_OBJECT('颜色', '银色', '容量', '128GB', '网络版本', '全网通版'),
    sort = 1,
    sales = 300,
    publish_status = 1,
    verify_status = 1
WHERE id = 1;

UPDATE pms_product_sku
SET name = '小米手机 银色 256GB 全网通版',
    sku_code = 'XM-256-SLV',
    price = 8999.00,
    promotion_price = 8899.00,
    stock = 800,
    low_stock = 80,
    spec_data = JSON_OBJECT('颜色', '银色', '容量', '256GB', '网络版本', '全网通版'),
    sort = 2,
    sales = 200,
    publish_status = 1,
    verify_status = 1
WHERE id = 2;

INSERT INTO pms_product_sku
    (platform_id, tenant_id, merchant_id, spu_id, name, sku_code, main_pic, album_pics, price, promotion_price, stock, low_stock, spec_data, weight, publish_status, verify_status, sort, sales)
SELECT 1, 0, 0, 1, '小米手机 金色 128GB 全网通版', 'XM-128-GLD', 'http://129.204.203.29/xiaomi_s.jpg', '', 7999.00, 7899.00, 920, 90,
       JSON_OBJECT('颜色', '金色', '容量', '128GB', '网络版本', '全网通版'), 0.19, 1, 1, 3, 260
WHERE NOT EXISTS (
    SELECT 1 FROM pms_product_sku WHERE sku_code = 'XM-128-GLD' AND is_deleted = 0
);

INSERT INTO pms_product_sku
    (platform_id, tenant_id, merchant_id, spu_id, name, sku_code, main_pic, album_pics, price, promotion_price, stock, low_stock, spec_data, weight, publish_status, verify_status, sort, sales)
SELECT 1, 0, 0, 1, '小米手机 金色 256GB 全网通版', 'XM-256-GLD', 'http://129.204.203.29/xiaomi_s.jpg', '', 8999.00, 8899.00, 760, 76,
       JSON_OBJECT('颜色', '金色', '容量', '256GB', '网络版本', '全网通版'), 0.19, 1, 1, 4, 180
WHERE NOT EXISTS (
    SELECT 1 FROM pms_product_sku WHERE sku_code = 'XM-256-GLD' AND is_deleted = 0
);

UPDATE pms_product_sku
SET name = '苹果手机 暗紫色 128GB 全网通版',
    sku_code = 'IPH17-128-DP',
    price = 7999.00,
    promotion_price = 7899.00,
    stock = 1000,
    low_stock = 100,
    spec_data = JSON_OBJECT('颜色', '暗紫色', '容量', '128GB', '网络版本', '全网通版'),
    sort = 1,
    sales = 300,
    publish_status = 1,
    verify_status = 1
WHERE id = 3;

UPDATE pms_product_sku
SET name = '苹果手机 暗紫色 256GB 全网通版',
    sku_code = 'IPH17-256-DP',
    price = 8999.00,
    promotion_price = 8899.00,
    stock = 800,
    low_stock = 80,
    spec_data = JSON_OBJECT('颜色', '暗紫色', '容量', '256GB', '网络版本', '全网通版'),
    sort = 2,
    sales = 200,
    publish_status = 1,
    verify_status = 1
WHERE id = 4;

INSERT INTO pms_product_sku
    (platform_id, tenant_id, merchant_id, spu_id, name, sku_code, main_pic, album_pics, price, promotion_price, stock, low_stock, spec_data, weight, publish_status, verify_status, sort, sales)
SELECT 1, 0, 0, 2, '苹果手机 自然色 128GB 全网通版', 'IPH17-128-NAT', 'http://129.204.203.29/apple_s.jpg', '', 7999.00, 7899.00, 930, 93,
       JSON_OBJECT('颜色', '自然色', '容量', '128GB', '网络版本', '全网通版'), 0.19, 1, 1, 3, 260
WHERE NOT EXISTS (
    SELECT 1 FROM pms_product_sku WHERE sku_code = 'IPH17-128-NAT' AND is_deleted = 0
);

INSERT INTO pms_product_sku
    (platform_id, tenant_id, merchant_id, spu_id, name, sku_code, main_pic, album_pics, price, promotion_price, stock, low_stock, spec_data, weight, publish_status, verify_status, sort, sales)
SELECT 1, 0, 0, 2, '苹果手机 自然色 256GB 全网通版', 'IPH17-256-NAT', 'http://129.204.203.29/apple_s.jpg', '', 8999.00, 8899.00, 770, 77,
       JSON_OBJECT('颜色', '自然色', '容量', '256GB', '网络版本', '全网通版'), 0.19, 1, 1, 4, 180
WHERE NOT EXISTS (
    SELECT 1 FROM pms_product_sku WHERE sku_code = 'IPH17-256-NAT' AND is_deleted = 0
);

UPDATE pms_product_sku
SET name = '华为手机 自然色 128GB 全网通版',
    sku_code = 'HW-128-NAT',
    price = 7999.00,
    promotion_price = 7899.00,
    stock = 1000,
    low_stock = 100,
    spec_data = JSON_OBJECT('颜色', '自然色', '容量', '128GB', '网络版本', '全网通版'),
    sort = 1,
    sales = 300,
    publish_status = 1,
    verify_status = 1
WHERE id = 5;

UPDATE pms_product_sku
SET name = '华为手机 自然色 256GB 全网通版',
    sku_code = 'HW-256-NAT',
    price = 8999.00,
    promotion_price = 8899.00,
    stock = 800,
    low_stock = 80,
    spec_data = JSON_OBJECT('颜色', '自然色', '容量', '256GB', '网络版本', '全网通版'),
    sort = 2,
    sales = 200,
    publish_status = 1,
    verify_status = 1
WHERE id = 6;

INSERT INTO pms_product_sku
    (platform_id, tenant_id, merchant_id, spu_id, name, sku_code, main_pic, album_pics, price, promotion_price, stock, low_stock, spec_data, weight, publish_status, verify_status, sort, sales)
SELECT 1, 0, 0, 3, '华为手机 银色 128GB 全网通版', 'HW-128-SLV', 'http://129.204.203.29/hua_s.jpg', '', 7999.00, 7899.00, 900, 90,
       JSON_OBJECT('颜色', '银色', '容量', '128GB', '网络版本', '全网通版'), 0.19, 1, 1, 3, 250
WHERE NOT EXISTS (
    SELECT 1 FROM pms_product_sku WHERE sku_code = 'HW-128-SLV' AND is_deleted = 0
);

INSERT INTO pms_product_sku
    (platform_id, tenant_id, merchant_id, spu_id, name, sku_code, main_pic, album_pics, price, promotion_price, stock, low_stock, spec_data, weight, publish_status, verify_status, sort, sales)
SELECT 1, 0, 0, 3, '华为手机 银色 256GB 全网通版', 'HW-256-SLV', 'http://129.204.203.29/hua_s.jpg', '', 8999.00, 8899.00, 750, 75,
       JSON_OBJECT('颜色', '银色', '容量', '256GB', '网络版本', '全网通版'), 0.19, 1, 1, 4, 170
WHERE NOT EXISTS (
    SELECT 1 FROM pms_product_sku WHERE sku_code = 'HW-256-SLV' AND is_deleted = 0
);

UPDATE pms_product_sku
SET name = '三星手机 银色 128GB 全网通版',
    sku_code = 'SS-128-SLV',
    price = 7999.00,
    promotion_price = 7899.00,
    stock = 1000,
    low_stock = 100,
    spec_data = JSON_OBJECT('颜色', '银色', '容量', '128GB', '网络版本', '全网通版'),
    sort = 1,
    sales = 300,
    publish_status = 1,
    verify_status = 1
WHERE id = 7;

UPDATE pms_product_sku
SET name = '三星手机 银色 256GB 全网通版',
    sku_code = 'SS-256-SLV',
    price = 8999.00,
    promotion_price = 8899.00,
    stock = 800,
    low_stock = 80,
    spec_data = JSON_OBJECT('颜色', '银色', '容量', '256GB', '网络版本', '全网通版'),
    sort = 2,
    sales = 200,
    publish_status = 1,
    verify_status = 1
WHERE id = 8;

INSERT INTO pms_product_sku
    (platform_id, tenant_id, merchant_id, spu_id, name, sku_code, main_pic, album_pics, price, promotion_price, stock, low_stock, spec_data, weight, publish_status, verify_status, sort, sales)
SELECT 1, 0, 0, 4, '三星手机 暗紫色 128GB 全网通版', 'SS-128-DP', 'http://129.204.203.29/sumsang.jpg', '', 7999.00, 7899.00, 910, 91,
       JSON_OBJECT('颜色', '暗紫色', '容量', '128GB', '网络版本', '全网通版'), 0.19, 1, 1, 3, 240
WHERE NOT EXISTS (
    SELECT 1 FROM pms_product_sku WHERE sku_code = 'SS-128-DP' AND is_deleted = 0
);

INSERT INTO pms_product_sku
    (platform_id, tenant_id, merchant_id, spu_id, name, sku_code, main_pic, album_pics, price, promotion_price, stock, low_stock, spec_data, weight, publish_status, verify_status, sort, sales)
SELECT 1, 0, 0, 4, '三星手机 暗紫色 256GB 全网通版', 'SS-256-DP', 'http://129.204.203.29/sumsang.jpg', '', 8999.00, 8899.00, 760, 76,
       JSON_OBJECT('颜色', '暗紫色', '容量', '256GB', '网络版本', '全网通版'), 0.19, 1, 1, 4, 160
WHERE NOT EXISTS (
    SELECT 1 FROM pms_product_sku WHERE sku_code = 'SS-256-DP' AND is_deleted = 0
);

UPDATE pms_product_sku
SET name = '荣耀手机 金色 128GB 全网通版',
    sku_code = 'HON-128-GLD',
    price = 7999.00,
    promotion_price = 7899.00,
    stock = 1000,
    low_stock = 100,
    spec_data = JSON_OBJECT('颜色', '金色', '容量', '128GB', '网络版本', '全网通版'),
    sort = 1,
    sales = 300,
    publish_status = 1,
    verify_status = 1
WHERE id = 9;

UPDATE pms_product_sku
SET name = '荣耀手机 金色 256GB 全网通版',
    sku_code = 'HON-256-GLD',
    price = 8999.00,
    promotion_price = 8899.00,
    stock = 800,
    low_stock = 80,
    spec_data = JSON_OBJECT('颜色', '金色', '容量', '256GB', '网络版本', '全网通版'),
    sort = 2,
    sales = 200,
    publish_status = 1,
    verify_status = 1
WHERE id = 10;

INSERT INTO pms_product_sku
    (platform_id, tenant_id, merchant_id, spu_id, name, sku_code, main_pic, album_pics, price, promotion_price, stock, low_stock, spec_data, weight, publish_status, verify_status, sort, sales)
SELECT 1, 0, 0, 5, '荣耀手机 自然色 128GB 全网通版', 'HON-128-NAT', 'http://129.204.203.29/rong_s.jpg', '', 7999.00, 7899.00, 910, 91,
       JSON_OBJECT('颜色', '自然色', '容量', '128GB', '网络版本', '全网通版'), 0.19, 1, 1, 3, 250
WHERE NOT EXISTS (
    SELECT 1 FROM pms_product_sku WHERE sku_code = 'HON-128-NAT' AND is_deleted = 0
);

INSERT INTO pms_product_sku
    (platform_id, tenant_id, merchant_id, spu_id, name, sku_code, main_pic, album_pics, price, promotion_price, stock, low_stock, spec_data, weight, publish_status, verify_status, sort, sales)
SELECT 1, 0, 0, 5, '荣耀手机 自然色 256GB 全网通版', 'HON-256-NAT', 'http://129.204.203.29/rong_s.jpg', '', 8999.00, 8899.00, 760, 76,
       JSON_OBJECT('颜色', '自然色', '容量', '256GB', '网络版本', '全网通版'), 0.19, 1, 1, 4, 170
WHERE NOT EXISTS (
    SELECT 1 FROM pms_product_sku WHERE sku_code = 'HON-256-NAT' AND is_deleted = 0
);

COMMIT;
