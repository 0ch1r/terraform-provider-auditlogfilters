package provider

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestConnectAndVerifyOpenError(t *testing.T) {
	originalOpen := sqlOpenFunc
	originalPing := pingDBFunc
	originalQuery := queryComponentCountFunc
	t.Cleanup(func() {
		sqlOpenFunc = originalOpen
		pingDBFunc = originalPing
		queryComponentCountFunc = originalQuery
	})

	sqlOpenFunc = func(driverName, dataSourceName string) (*sql.DB, error) {
		return nil, errors.New("open failed")
	}

	var diagnostics diag.Diagnostics
	_, ok := connectAndVerify(context.Background(), providerValidatedConfig{}, &diagnostics)
	if ok {
		t.Fatalf("expected connectAndVerify to fail on open error")
	}
	if !diagnostics.HasError() {
		t.Fatalf("expected diagnostics error for open failure")
	}
	if diagnostics[0].Summary() != "Unable to Create MySQL Client" {
		t.Fatalf("unexpected diagnostic summary: %q", diagnostics[0].Summary())
	}
}

func TestConnectAndVerifyPingError(t *testing.T) {
	originalOpen := sqlOpenFunc
	originalPing := pingDBFunc
	originalQuery := queryComponentCountFunc
	t.Cleanup(func() {
		sqlOpenFunc = originalOpen
		pingDBFunc = originalPing
		queryComponentCountFunc = originalQuery
	})

	sqlOpenFunc = func(driverName, dataSourceName string) (*sql.DB, error) {
		return sql.Open("mysql", "")
	}
	pingDBFunc = func(ctx context.Context, db *sql.DB) error {
		return errors.New("ping failed")
	}

	var diagnostics diag.Diagnostics
	_, ok := connectAndVerify(context.Background(), providerValidatedConfig{}, &diagnostics)
	if ok {
		t.Fatalf("expected connectAndVerify to fail on ping error")
	}
	if !diagnostics.HasError() {
		t.Fatalf("expected diagnostics error for ping failure")
	}
	if diagnostics[0].Summary() != "Unable to Connect to MySQL" {
		t.Fatalf("unexpected diagnostic summary: %q", diagnostics[0].Summary())
	}
}

func TestConnectAndVerifyComponentMissing(t *testing.T) {
	originalOpen := sqlOpenFunc
	originalPing := pingDBFunc
	originalQuery := queryComponentCountFunc
	t.Cleanup(func() {
		sqlOpenFunc = originalOpen
		pingDBFunc = originalPing
		queryComponentCountFunc = originalQuery
	})

	sqlOpenFunc = func(driverName, dataSourceName string) (*sql.DB, error) {
		return sql.Open("mysql", "")
	}
	pingDBFunc = func(ctx context.Context, db *sql.DB) error {
		return nil
	}
	queryComponentCountFunc = func(ctx context.Context, db *sql.DB) (int, error) {
		return 0, nil
	}

	var diagnostics diag.Diagnostics
	_, ok := connectAndVerify(context.Background(), providerValidatedConfig{}, &diagnostics)
	if ok {
		t.Fatalf("expected connectAndVerify to fail when component is missing")
	}
	if !diagnostics.HasError() {
		t.Fatalf("expected diagnostics error for missing component")
	}
	if diagnostics[0].Summary() != "Audit Log Filter Component Not Available" {
		t.Fatalf("unexpected diagnostic summary: %q", diagnostics[0].Summary())
	}
}

func TestConnectAndVerifySuccess(t *testing.T) {
	originalOpen := sqlOpenFunc
	originalPing := pingDBFunc
	originalQuery := queryComponentCountFunc
	t.Cleanup(func() {
		sqlOpenFunc = originalOpen
		pingDBFunc = originalPing
		queryComponentCountFunc = originalQuery
	})

	sqlOpenFunc = func(driverName, dataSourceName string) (*sql.DB, error) {
		return sql.Open("mysql", "")
	}
	pingDBFunc = func(ctx context.Context, db *sql.DB) error {
		return nil
	}
	queryComponentCountFunc = func(ctx context.Context, db *sql.DB) (int, error) {
		return 1, nil
	}

	validated := providerValidatedConfig{
		maxLifetime:  2 * time.Minute,
		maxOpenConns: 9,
		maxIdleConns: 4,
	}

	var diagnostics diag.Diagnostics
	db, ok := connectAndVerify(context.Background(), validated, &diagnostics)
	if !ok {
		t.Fatalf("expected connectAndVerify to succeed, diagnostics: %+v", diagnostics)
	}
	if diagnostics.HasError() {
		t.Fatalf("expected no diagnostics errors, got: %+v", diagnostics)
	}
	if db == nil {
		t.Fatalf("expected db to be returned")
	}
	_ = db.Close()
}
