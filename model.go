package main

import (
	"database/sql"

	"github.com/google/uuid"
)

type product struct {
	ID          string `json:"id"`
	SKU         string `json:"sku"`
	ProductName string `json:"product_name"`
}

func getProducts(db *sql.DB, limit int, offset int) ([]product, error) {
	rows, err := db.Query("SELECT id, sku, product_name FROM products LIMIT ? OFFSET ?",
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []product{}

	for rows.Next() {
		var p product
		if err := rows.Scan(&p.ID, &p.SKU, &p.ProductName); err != nil {
			return nil, err
		}

		products = append(products, p)
	}

	return products, nil
}

func (p *product) getProductBySKU(db *sql.DB) error {
	err := db.QueryRow("SELECT id, sku, product_name FROM products WHERE sku=?",
		p.SKU).Scan(&p.ID, &p.SKU, &p.ProductName)
	if err != nil {
		return err
	}

	return nil
}

func (p *product) createProduct(db *sql.DB) error {
	p.ID = uuid.New().String()

	_, err := db.Exec("INSERT INTO products(id, sku, product_name) VALUES(?, ?, ?)", p.ID, p.SKU, p.ProductName)
	if err != nil {
		return err
	}

	return db.QueryRow("SELECT id, sku, product_name FROM products WHERE sku=?",
		p.SKU).Scan(&p.ID, &p.SKU, &p.ProductName)
}

func (p *product) deleteProduct(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM products WHERE sku=?", p.SKU)
	if err != nil {
		return err
	}

	return nil
}
