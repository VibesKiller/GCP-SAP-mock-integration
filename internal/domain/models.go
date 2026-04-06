package domain

import (
	"errors"
	"fmt"
	"strings"
)

type SalesOrderTotals struct {
	NetAmount   float64 `json:"net_amount"`
	TaxAmount   float64 `json:"tax_amount"`
	TotalAmount float64 `json:"total_amount"`
}

type SAPSalesOrderItem struct {
	LineNumber   int     `json:"line_number"`
	MaterialCode string  `json:"material_code"`
	Description  string  `json:"description"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	UnitPrice    float64 `json:"unit_price"`
	NetAmount    float64 `json:"net_amount"`
}

type SAPSalesOrderPayload struct {
	SalesDocumentID       string              `json:"sales_document_id"`
	SalesOrganization     string              `json:"sales_organization"`
	DistributionChannel   string              `json:"distribution_channel"`
	Division              string              `json:"division"`
	SoldToPartyID         string              `json:"sold_to_party_id"`
	Currency              string              `json:"currency"`
	RequestedDeliveryDate string              `json:"requested_delivery_date"`
	DocumentDate          string              `json:"document_date"`
	Status                string              `json:"status"`
	NetAmount             float64             `json:"net_amount"`
	TaxAmount             float64             `json:"tax_amount"`
	TotalAmount           float64             `json:"total_amount"`
	Items                 []SAPSalesOrderItem `json:"items"`
}

type SAPCustomerPayload struct {
	CustomerID     string `json:"customer_id"`
	CustomerNumber string `json:"customer_number"`
	FullName       string `json:"full_name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	CountryCode    string `json:"country_code"`
	City           string `json:"city"`
	PostalCode     string `json:"postal_code"`
	Segment        string `json:"segment"`
	Status         string `json:"status"`
}

type SAPInvoicePayload struct {
	BillingDocumentID string  `json:"billing_document_id"`
	SalesDocumentID   string  `json:"sales_document_id"`
	CustomerID        string  `json:"customer_id"`
	Currency          string  `json:"currency"`
	IssueDate         string  `json:"issue_date"`
	DueDate           string  `json:"due_date"`
	Status            string  `json:"status"`
	NetAmount         float64 `json:"net_amount"`
	TaxAmount         float64 `json:"tax_amount"`
	TotalAmount       float64 `json:"total_amount"`
}

type SalesOrderItem struct {
	LineNumber  int     `json:"line_number"`
	SKU         string  `json:"sku"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unit_price"`
	NetAmount   float64 `json:"net_amount"`
}

type SalesOrderPayload struct {
	SalesOrderID          string           `json:"sales_order_id"`
	CustomerID            string           `json:"customer_id"`
	SalesOrg              string           `json:"sales_org"`
	DistributionChannel   string           `json:"distribution_channel"`
	Division              string           `json:"division"`
	Currency              string           `json:"currency"`
	RequestedDeliveryDate string           `json:"requested_delivery_date"`
	DocumentDate          string           `json:"document_date"`
	Status                string           `json:"status"`
	Totals                SalesOrderTotals `json:"totals"`
	Items                 []SalesOrderItem `json:"items"`
}

type CustomerPayload struct {
	CustomerID     string `json:"customer_id"`
	CustomerNumber string `json:"customer_number"`
	FullName       string `json:"full_name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	CountryCode    string `json:"country_code"`
	City           string `json:"city"`
	PostalCode     string `json:"postal_code"`
	Segment        string `json:"segment"`
	Status         string `json:"status"`
}

type InvoicePayload struct {
	InvoiceID    string  `json:"invoice_id"`
	SalesOrderID string  `json:"sales_order_id"`
	CustomerID   string  `json:"customer_id"`
	Currency     string  `json:"currency"`
	IssueDate    string  `json:"issue_date"`
	DueDate      string  `json:"due_date"`
	Status       string  `json:"status"`
	NetAmount    float64 `json:"net_amount"`
	TaxAmount    float64 `json:"tax_amount"`
	TotalAmount  float64 `json:"total_amount"`
}

func (p SAPSalesOrderPayload) Validate() error {
	switch {
	case strings.TrimSpace(p.SalesDocumentID) == "":
		return errors.New("sales_document_id is required")
	case strings.TrimSpace(p.SalesOrganization) == "":
		return errors.New("sales_organization is required")
	case strings.TrimSpace(p.DistributionChannel) == "":
		return errors.New("distribution_channel is required")
	case strings.TrimSpace(p.Division) == "":
		return errors.New("division is required")
	case strings.TrimSpace(p.SoldToPartyID) == "":
		return errors.New("sold_to_party_id is required")
	case strings.TrimSpace(p.Currency) == "":
		return errors.New("currency is required")
	case strings.TrimSpace(p.Status) == "":
		return errors.New("status is required")
	case strings.TrimSpace(p.RequestedDeliveryDate) == "":
		return errors.New("requested_delivery_date is required")
	case strings.TrimSpace(p.DocumentDate) == "":
		return errors.New("document_date is required")
	case len(p.Items) == 0:
		return errors.New("at least one order item is required")
	}

	for _, item := range p.Items {
		switch {
		case item.LineNumber <= 0:
			return fmt.Errorf("item line_number must be positive")
		case strings.TrimSpace(item.MaterialCode) == "":
			return fmt.Errorf("item material_code is required")
		case strings.TrimSpace(item.Description) == "":
			return fmt.Errorf("item description is required")
		case item.Quantity <= 0:
			return fmt.Errorf("item quantity must be positive")
		case strings.TrimSpace(item.Unit) == "":
			return fmt.Errorf("item unit is required")
		case item.UnitPrice < 0:
			return fmt.Errorf("item unit_price must be non-negative")
		}
	}

	return nil
}

func (p SAPCustomerPayload) Validate() error {
	switch {
	case strings.TrimSpace(p.CustomerID) == "":
		return errors.New("customer_id is required")
	case strings.TrimSpace(p.CustomerNumber) == "":
		return errors.New("customer_number is required")
	case strings.TrimSpace(p.FullName) == "":
		return errors.New("full_name is required")
	case strings.TrimSpace(p.Email) == "":
		return errors.New("email is required")
	case strings.TrimSpace(p.CountryCode) == "":
		return errors.New("country_code is required")
	case strings.TrimSpace(p.Status) == "":
		return errors.New("status is required")
	}

	return nil
}

func (p SAPInvoicePayload) Validate() error {
	switch {
	case strings.TrimSpace(p.BillingDocumentID) == "":
		return errors.New("billing_document_id is required")
	case strings.TrimSpace(p.SalesDocumentID) == "":
		return errors.New("sales_document_id is required")
	case strings.TrimSpace(p.CustomerID) == "":
		return errors.New("customer_id is required")
	case strings.TrimSpace(p.Currency) == "":
		return errors.New("currency is required")
	case strings.TrimSpace(p.IssueDate) == "":
		return errors.New("issue_date is required")
	case strings.TrimSpace(p.DueDate) == "":
		return errors.New("due_date is required")
	case strings.TrimSpace(p.Status) == "":
		return errors.New("status is required")
	}

	return nil
}

func NormalizeSalesOrderPayload(input SAPSalesOrderPayload) SalesOrderPayload {
	items := make([]SalesOrderItem, 0, len(input.Items))
	for _, item := range input.Items {
		items = append(items, SalesOrderItem{
			LineNumber:  item.LineNumber,
			SKU:         item.MaterialCode,
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			NetAmount:   item.NetAmount,
		})
	}

	return SalesOrderPayload{
		SalesOrderID:          input.SalesDocumentID,
		CustomerID:            input.SoldToPartyID,
		SalesOrg:              input.SalesOrganization,
		DistributionChannel:   input.DistributionChannel,
		Division:              input.Division,
		Currency:              input.Currency,
		RequestedDeliveryDate: input.RequestedDeliveryDate,
		DocumentDate:          input.DocumentDate,
		Status:                input.Status,
		Totals: SalesOrderTotals{
			NetAmount:   input.NetAmount,
			TaxAmount:   input.TaxAmount,
			TotalAmount: input.TotalAmount,
		},
		Items: items,
	}
}

func NormalizeCustomerPayload(input SAPCustomerPayload) CustomerPayload {
	return CustomerPayload{
		CustomerID:     input.CustomerID,
		CustomerNumber: input.CustomerNumber,
		FullName:       input.FullName,
		Email:          input.Email,
		Phone:          input.Phone,
		CountryCode:    input.CountryCode,
		City:           input.City,
		PostalCode:     input.PostalCode,
		Segment:        input.Segment,
		Status:         input.Status,
	}
}

func NormalizeInvoicePayload(input SAPInvoicePayload) InvoicePayload {
	return InvoicePayload{
		InvoiceID:    input.BillingDocumentID,
		SalesOrderID: input.SalesDocumentID,
		CustomerID:   input.CustomerID,
		Currency:     input.Currency,
		IssueDate:    input.IssueDate,
		DueDate:      input.DueDate,
		Status:       input.Status,
		NetAmount:    input.NetAmount,
		TaxAmount:    input.TaxAmount,
		TotalAmount:  input.TotalAmount,
	}
}
