package main

import (
	"net/http/httptest"
	"testing"
)

func TestParsePaginationDefaults(t *testing.T) {
	application := &app{config: appConfig{DefaultPageSize: 20, MaxPageSize: 100}}
	request := httptest.NewRequest("GET", "/api/v1/orders", nil)

	limit, offset, err := application.parsePagination(request)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if limit != 20 || offset != 0 {
		t.Fatalf("expected limit=20 and offset=0, got limit=%d offset=%d", limit, offset)
	}
}

func TestParsePaginationRejectsTooLargeLimit(t *testing.T) {
	application := &app{config: appConfig{DefaultPageSize: 20, MaxPageSize: 100}}
	request := httptest.NewRequest("GET", "/api/v1/orders?limit=101", nil)

	if _, _, err := application.parsePagination(request); err == nil {
		t.Fatal("expected error for limit above max page size")
	}
}
