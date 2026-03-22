create table cms_subject_product_relation
(
    id          bigint auto_increment
        primary key comment '主键ID',
    platform_id bigint default 1 not null comment '平台ID',
    tenant_id   bigint default 0 not null comment '租户ID',
    merchant_id bigint default 0 not null comment '商户ID',
    subject_id  bigint not null comment '专题ID',
    product_id  bigint not null comment '商品ID'
)
    comment '专题商品关系表' charset = utf8;

create index idx_subject_relation_scope_subject
    on cms_subject_product_relation (platform_id, tenant_id, merchant_id, subject_id);

create index idx_subject_relation_scope_product
    on cms_subject_product_relation (platform_id, tenant_id, merchant_id, product_id);

INSERT INTO cms_subject_product_relation (id, subject_id, product_id) VALUES (1, 1, 12);
INSERT INTO cms_subject_product_relation (id, subject_id, product_id) VALUES (2, 1, 13);
INSERT INTO cms_subject_product_relation (id, subject_id, product_id) VALUES (3, 1, 14);
INSERT INTO cms_subject_product_relation (id, subject_id, product_id) VALUES (4, 1, 18);
INSERT INTO cms_subject_product_relation (id, subject_id, product_id) VALUES (5, 1, 7);
INSERT INTO cms_subject_product_relation (id, subject_id, product_id) VALUES (6, 2, 7);
INSERT INTO cms_subject_product_relation (id, subject_id, product_id) VALUES (7, 1, 22);
INSERT INTO cms_subject_product_relation (id, subject_id, product_id) VALUES (29, 1, 23);
INSERT INTO cms_subject_product_relation (id, subject_id, product_id) VALUES (30, 4, 23);
INSERT INTO cms_subject_product_relation (id, subject_id, product_id) VALUES (31, 5, 23);
INSERT INTO cms_subject_product_relation (id, subject_id, product_id) VALUES (68, 2, 26);
INSERT INTO cms_subject_product_relation (id, subject_id, product_id) VALUES (69, 3, 26);
INSERT INTO cms_subject_product_relation (id, subject_id, product_id) VALUES (70, 6, 26);
