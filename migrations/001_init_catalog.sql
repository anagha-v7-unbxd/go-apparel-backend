-- Catalog schema for go-starter product API
-- Run this against your PostgreSQL database before using ingest/browse/delete.

CREATE TABLE IF NOT EXISTS catalog_data (
    unique_id          VARCHAR(255) PRIMARY KEY,
    sku                VARCHAR(255),
    name               TEXT,
    title              TEXT,
    product_description TEXT,
    product_url        TEXT,
    product_image      TEXT,
    price              DOUBLE PRECISION,
    availability       VARCHAR(100),
    category_type      VARCHAR(255),
    catlevel1_name     VARCHAR(255),
    catlevel2_name     VARCHAR(255),
    catlevel3_name     VARCHAR(255),
    catlevel4_name     VARCHAR(255),
    color              TEXT,   -- JSON array as text
    size               TEXT,   -- JSON array as text
    category           TEXT,   -- JSON array as text
    gender             TEXT    -- JSON array as text
);

CREATE INDEX IF NOT EXISTS idx_catalog_catlevel1 ON catalog_data(catlevel1_name);
CREATE INDEX IF NOT EXISTS idx_catalog_catlevel2 ON catalog_data(catlevel2_name);
CREATE INDEX IF NOT EXISTS idx_catalog_catlevel3 ON catalog_data(catlevel3_name);
CREATE INDEX IF NOT EXISTS idx_catalog_category_type ON catalog_data(category_type);
CREATE INDEX IF NOT EXISTS idx_catalog_price ON catalog_data(price);
