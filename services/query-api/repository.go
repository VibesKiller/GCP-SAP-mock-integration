package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

type pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type listResponse[T any] struct {
	Data       []T        `json:"data"`
	Pagination pagination `json:"pagination"`
}

type customer struct {
	CustomerID        string    `json:"customer_id"`
	CustomerNumber    string    `json:"customer_number"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	Phone             string    `json:"phone"`
	CountryCode       string    `json:"country_code"`
	City              string    `json:"city"`
	PostalCode        string    `json:"postal_code"`
	Segment           string    `json:"segment"`
	Status            string    `json:"status"`
	SourceUpdatedAt   time.Time `json:"source_updated_at"`
	LastEventID       string    `json:"last_event_id"`
	LastCorrelationID string    `json:"last_correlation_id"`
}

type orderItem struct {
	LineNumber  int     `json:"line_number"`
	SKU         string  `json:"sku"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unit_price"`
	NetAmount   float64 `json:"net_amount"`
}

type orderSummary struct {
	OrderID               string    `json:"order_id"`
	CustomerID            string    `json:"customer_id"`
	SalesOrg              string    `json:"sales_org"`
	DistributionChannel   string    `json:"distribution_channel"`
	Division              string    `json:"division"`
	Currency              string    `json:"currency"`
	Status                string    `json:"status"`
	RequestedDeliveryDate string    `json:"requested_delivery_date"`
	DocumentDate          string    `json:"document_date"`
	NetAmount             float64   `json:"net_amount"`
	TaxAmount             float64   `json:"tax_amount"`
	TotalAmount           float64   `json:"total_amount"`
	SourceUpdatedAt       time.Time `json:"source_updated_at"`
	LastEventID           string    `json:"last_event_id"`
	LastCorrelationID     string    `json:"last_correlation_id"`
}

type orderDetail struct {
	orderSummary
	Items []orderItem `json:"items"`
}

type invoice struct {
	InvoiceID         string    `json:"invoice_id"`
	OrderID           string    `json:"order_id"`
	CustomerID        string    `json:"customer_id"`
	Currency          string    `json:"currency"`
	Status            string    `json:"status"`
	IssueDate         time.Time `json:"issue_date"`
	DueDate           time.Time `json:"due_date"`
	NetAmount         float64   `json:"net_amount"`
	TaxAmount         float64   `json:"tax_amount"`
	TotalAmount       float64   `json:"total_amount"`
	SourceUpdatedAt   time.Time `json:"source_updated_at"`
	LastEventID       string    `json:"last_event_id"`
	LastCorrelationID string    `json:"last_correlation_id"`
}

type customerFilters struct {
	Status      string
	CountryCode string
	Segment     string
}

type orderFilters struct {
	CustomerID string
	Status     string
}

type invoiceFilters struct {
	CustomerID string
	OrderID    string
	Status     string
}

func newRepository(db *pgxpool.Pool) *repository {
	return &repository{db: db}
}

func (r *repository) ready(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func (r *repository) listCustomers(ctx context.Context, limit, offset int, filters customerFilters) (listResponse[customer], error) {
	clauses, args := buildCustomerFilters(filters)

	total, err := r.count(ctx, "customers", clauses, args)
	if err != nil {
		return listResponse[customer]{}, err
	}

	dataArgs := append(append([]any{}, args...), limit, offset)
	query := fmt.Sprintf(`
    SELECT customer_id, customer_number, full_name, email, phone, country_code, city, postal_code, segment, status,
           source_updated_at, last_event_id, last_correlation_id
    FROM customers
    %s
    ORDER BY updated_at DESC
    LIMIT $%d OFFSET $%d
  `, whereClause(clauses), len(args)+1, len(args)+2)

	rows, err := r.db.Query(ctx, query, dataArgs...)
	if err != nil {
		return listResponse[customer]{}, err
	}
	defer rows.Close()

	customers := make([]customer, 0, limit)
	for rows.Next() {
		var item customer
		if err := rows.Scan(
			&item.CustomerID,
			&item.CustomerNumber,
			&item.FullName,
			&item.Email,
			&item.Phone,
			&item.CountryCode,
			&item.City,
			&item.PostalCode,
			&item.Segment,
			&item.Status,
			&item.SourceUpdatedAt,
			&item.LastEventID,
			&item.LastCorrelationID,
		); err != nil {
			return listResponse[customer]{}, err
		}
		customers = append(customers, item)
	}

	return listResponse[customer]{
		Data:       customers,
		Pagination: pagination{Limit: limit, Offset: offset, Total: total},
	}, rows.Err()
}

func (r *repository) getCustomer(ctx context.Context, customerID string) (customer, error) {
	var item customer
	err := r.db.QueryRow(ctx, `
    SELECT customer_id, customer_number, full_name, email, phone, country_code, city, postal_code, segment, status,
           source_updated_at, last_event_id, last_correlation_id
    FROM customers
    WHERE customer_id = $1
  `, customerID).Scan(
		&item.CustomerID,
		&item.CustomerNumber,
		&item.FullName,
		&item.Email,
		&item.Phone,
		&item.CountryCode,
		&item.City,
		&item.PostalCode,
		&item.Segment,
		&item.Status,
		&item.SourceUpdatedAt,
		&item.LastEventID,
		&item.LastCorrelationID,
	)
	return item, err
}

func (r *repository) listOrders(ctx context.Context, limit, offset int, filters orderFilters) (listResponse[orderSummary], error) {
	clauses, args := buildOrderFilters(filters)
	total, err := r.count(ctx, "orders", clauses, args)
	if err != nil {
		return listResponse[orderSummary]{}, err
	}

	dataArgs := append(append([]any{}, args...), limit, offset)
	query := fmt.Sprintf(`
    SELECT order_id, customer_id, sales_org, distribution_channel, division, currency, status,
           requested_delivery_date::text, document_date::text, net_amount::float8, tax_amount::float8, total_amount::float8,
           source_updated_at, last_event_id, last_correlation_id
    FROM orders
    %s
    ORDER BY updated_at DESC
    LIMIT $%d OFFSET $%d
  `, whereClause(clauses), len(args)+1, len(args)+2)

	rows, err := r.db.Query(ctx, query, dataArgs...)
	if err != nil {
		return listResponse[orderSummary]{}, err
	}
	defer rows.Close()

	orders := make([]orderSummary, 0, limit)
	for rows.Next() {
		var item orderSummary
		if err := rows.Scan(
			&item.OrderID,
			&item.CustomerID,
			&item.SalesOrg,
			&item.DistributionChannel,
			&item.Division,
			&item.Currency,
			&item.Status,
			&item.RequestedDeliveryDate,
			&item.DocumentDate,
			&item.NetAmount,
			&item.TaxAmount,
			&item.TotalAmount,
			&item.SourceUpdatedAt,
			&item.LastEventID,
			&item.LastCorrelationID,
		); err != nil {
			return listResponse[orderSummary]{}, err
		}
		orders = append(orders, item)
	}

	return listResponse[orderSummary]{
		Data:       orders,
		Pagination: pagination{Limit: limit, Offset: offset, Total: total},
	}, rows.Err()
}

func (r *repository) getOrder(ctx context.Context, orderID string) (orderDetail, error) {
	var detail orderDetail
	err := r.db.QueryRow(ctx, `
    SELECT order_id, customer_id, sales_org, distribution_channel, division, currency, status,
           requested_delivery_date::text, document_date::text, net_amount::float8, tax_amount::float8, total_amount::float8,
           source_updated_at, last_event_id, last_correlation_id
    FROM orders
    WHERE order_id = $1
  `, orderID).Scan(
		&detail.OrderID,
		&detail.CustomerID,
		&detail.SalesOrg,
		&detail.DistributionChannel,
		&detail.Division,
		&detail.Currency,
		&detail.Status,
		&detail.RequestedDeliveryDate,
		&detail.DocumentDate,
		&detail.NetAmount,
		&detail.TaxAmount,
		&detail.TotalAmount,
		&detail.SourceUpdatedAt,
		&detail.LastEventID,
		&detail.LastCorrelationID,
	)
	if err != nil {
		return orderDetail{}, err
	}

	rows, err := r.db.Query(ctx, `
    SELECT line_number, sku, description, quantity::float8, unit, unit_price::float8, net_amount::float8
    FROM order_items
    WHERE order_id = $1
    ORDER BY line_number ASC
  `, orderID)
	if err != nil {
		return orderDetail{}, err
	}
	defer rows.Close()

	items := make([]orderItem, 0)
	for rows.Next() {
		var item orderItem
		if err := rows.Scan(
			&item.LineNumber,
			&item.SKU,
			&item.Description,
			&item.Quantity,
			&item.Unit,
			&item.UnitPrice,
			&item.NetAmount,
		); err != nil {
			return orderDetail{}, err
		}
		items = append(items, item)
	}

	detail.Items = items
	return detail, rows.Err()
}

func (r *repository) listInvoices(ctx context.Context, limit, offset int, filters invoiceFilters) (listResponse[invoice], error) {
	clauses, args := buildInvoiceFilters(filters)
	total, err := r.count(ctx, "invoices", clauses, args)
	if err != nil {
		return listResponse[invoice]{}, err
	}

	dataArgs := append(append([]any{}, args...), limit, offset)
	query := fmt.Sprintf(`
    SELECT invoice_id, order_id, customer_id, currency, status, issue_date, due_date,
           net_amount::float8, tax_amount::float8, total_amount::float8,
           source_updated_at, last_event_id, last_correlation_id
    FROM invoices
    %s
    ORDER BY issue_date DESC
    LIMIT $%d OFFSET $%d
  `, whereClause(clauses), len(args)+1, len(args)+2)

	rows, err := r.db.Query(ctx, query, dataArgs...)
	if err != nil {
		return listResponse[invoice]{}, err
	}
	defer rows.Close()

	invoices := make([]invoice, 0, limit)
	for rows.Next() {
		var item invoice
		if err := rows.Scan(
			&item.InvoiceID,
			&item.OrderID,
			&item.CustomerID,
			&item.Currency,
			&item.Status,
			&item.IssueDate,
			&item.DueDate,
			&item.NetAmount,
			&item.TaxAmount,
			&item.TotalAmount,
			&item.SourceUpdatedAt,
			&item.LastEventID,
			&item.LastCorrelationID,
		); err != nil {
			return listResponse[invoice]{}, err
		}
		invoices = append(invoices, item)
	}

	return listResponse[invoice]{
		Data:       invoices,
		Pagination: pagination{Limit: limit, Offset: offset, Total: total},
	}, rows.Err()
}

func (r *repository) getInvoice(ctx context.Context, invoiceID string) (invoice, error) {
	var item invoice
	err := r.db.QueryRow(ctx, `
    SELECT invoice_id, order_id, customer_id, currency, status, issue_date, due_date,
           net_amount::float8, tax_amount::float8, total_amount::float8,
           source_updated_at, last_event_id, last_correlation_id
    FROM invoices
    WHERE invoice_id = $1
  `, invoiceID).Scan(
		&item.InvoiceID,
		&item.OrderID,
		&item.CustomerID,
		&item.Currency,
		&item.Status,
		&item.IssueDate,
		&item.DueDate,
		&item.NetAmount,
		&item.TaxAmount,
		&item.TotalAmount,
		&item.SourceUpdatedAt,
		&item.LastEventID,
		&item.LastCorrelationID,
	)
	return item, err
}

func (r *repository) count(ctx context.Context, table string, clauses []string, args []any) (int, error) {
	query := fmt.Sprintf("SELECT count(*) FROM %s %s", table, whereClause(clauses))
	var total int
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func buildCustomerFilters(filters customerFilters) ([]string, []any) {
	clauses := make([]string, 0)
	args := make([]any, 0)
	if filters.Status != "" {
		args = append(args, filters.Status)
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)))
	}
	if filters.CountryCode != "" {
		args = append(args, filters.CountryCode)
		clauses = append(clauses, fmt.Sprintf("country_code = $%d", len(args)))
	}
	if filters.Segment != "" {
		args = append(args, filters.Segment)
		clauses = append(clauses, fmt.Sprintf("segment = $%d", len(args)))
	}
	return clauses, args
}

func buildOrderFilters(filters orderFilters) ([]string, []any) {
	clauses := make([]string, 0)
	args := make([]any, 0)
	if filters.CustomerID != "" {
		args = append(args, filters.CustomerID)
		clauses = append(clauses, fmt.Sprintf("customer_id = $%d", len(args)))
	}
	if filters.Status != "" {
		args = append(args, filters.Status)
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)))
	}
	return clauses, args
}

func buildInvoiceFilters(filters invoiceFilters) ([]string, []any) {
	clauses := make([]string, 0)
	args := make([]any, 0)
	if filters.CustomerID != "" {
		args = append(args, filters.CustomerID)
		clauses = append(clauses, fmt.Sprintf("customer_id = $%d", len(args)))
	}
	if filters.OrderID != "" {
		args = append(args, filters.OrderID)
		clauses = append(clauses, fmt.Sprintf("order_id = $%d", len(args)))
	}
	if filters.Status != "" {
		args = append(args, filters.Status)
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)))
	}
	return clauses, args
}

func whereClause(clauses []string) string {
	if len(clauses) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(clauses, " AND ")
}

func notFoundError(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
