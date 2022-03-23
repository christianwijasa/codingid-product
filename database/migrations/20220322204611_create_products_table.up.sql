CREATE TABLE IF NOT EXISTS products (
  id VARCHAR(255) PRIMARY KEY,
  sku VARCHAR(255) NOT NULL,
  product_name VARCHAR(255),
  UNIQUE(sku)
) ENGINE=INNODB;