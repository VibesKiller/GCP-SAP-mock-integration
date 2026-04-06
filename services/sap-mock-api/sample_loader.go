package main

import (
	"embed"
	"encoding/json"
	"fmt"

	"gcp-sap-mock-integration/internal/domain"
)

//go:embed sample-data/*.json
var sampleDataFS embed.FS

type sampleCatalog struct {
	SalesOrderCreate domain.SAPSalesOrderPayload `json:"sales_order_create"`
	SalesOrderUpdate domain.SAPSalesOrderPayload `json:"sales_order_update"`
	CustomerUpdate   domain.SAPCustomerPayload   `json:"customer_update"`
	InvoiceIssued    domain.SAPInvoicePayload    `json:"invoice_issued"`
}

func loadSamples() (sampleCatalog, error) {
	var catalog sampleCatalog

	if err := loadSampleJSON("sample-data/sales-order-create.json", &catalog.SalesOrderCreate); err != nil {
		return sampleCatalog{}, err
	}
	if err := loadSampleJSON("sample-data/sales-order-update.json", &catalog.SalesOrderUpdate); err != nil {
		return sampleCatalog{}, err
	}
	if err := loadSampleJSON("sample-data/customer-update.json", &catalog.CustomerUpdate); err != nil {
		return sampleCatalog{}, err
	}
	if err := loadSampleJSON("sample-data/invoice-issued.json", &catalog.InvoiceIssued); err != nil {
		return sampleCatalog{}, err
	}

	return catalog, nil
}

func loadSampleJSON(path string, target any) error {
	content, err := sampleDataFS.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read sample file %s: %w", path, err)
	}

	if err := json.Unmarshal(content, target); err != nil {
		return fmt.Errorf("decode sample file %s: %w", path, err)
	}

	return nil
}
