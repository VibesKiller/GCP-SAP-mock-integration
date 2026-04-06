package domain

import "testing"

func TestSAPSalesOrderPayloadValidateRequiresItems(t *testing.T) {
	payload := SAPSalesOrderPayload{
		SalesDocumentID:       "SO-2026-000184",
		SalesOrganization:     "CH10",
		DistributionChannel:   "B2B",
		Division:              "INDUSTRIAL",
		SoldToPartyID:         "CUST-100045",
		Currency:              "CHF",
		RequestedDeliveryDate: "2026-04-08",
		DocumentDate:          "2026-04-02",
		Status:                "OPEN",
	}

	if err := payload.Validate(); err == nil {
		t.Fatal("expected validation error when items are missing")
	}
}

func TestNormalizeSalesOrderPayload(t *testing.T) {
	payload := SAPSalesOrderPayload{
		SalesDocumentID:       "SO-2026-000184",
		SalesOrganization:     "CH10",
		DistributionChannel:   "B2B",
		Division:              "INDUSTRIAL",
		SoldToPartyID:         "CUST-100045",
		Currency:              "CHF",
		RequestedDeliveryDate: "2026-04-08",
		DocumentDate:          "2026-04-02",
		Status:                "OPEN",
		NetAmount:             12450,
		TaxAmount:             995.95,
		TotalAmount:           13445.95,
		Items: []SAPSalesOrderItem{{
			LineNumber:   10,
			MaterialCode: "MTR-VALVE-9000",
			Description:  "Industrial pressure valve 9000 series",
			Quantity:     15,
			Unit:         "EA",
			UnitPrice:    650,
			NetAmount:    9750,
		}},
	}

	normalized := NormalizeSalesOrderPayload(payload)
	if normalized.SalesOrderID != payload.SalesDocumentID {
		t.Fatalf("expected sales order ID %q, got %q", payload.SalesDocumentID, normalized.SalesOrderID)
	}
	if normalized.CustomerID != payload.SoldToPartyID {
		t.Fatalf("expected customer ID %q, got %q", payload.SoldToPartyID, normalized.CustomerID)
	}
	if normalized.Items[0].SKU != payload.Items[0].MaterialCode {
		t.Fatalf("expected SKU %q, got %q", payload.Items[0].MaterialCode, normalized.Items[0].SKU)
	}
	if normalized.Totals.TotalAmount != payload.TotalAmount {
		t.Fatalf("expected total amount %.2f, got %.2f", payload.TotalAmount, normalized.Totals.TotalAmount)
	}
}
