package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/vibast-solutions/ms-go-profile/app/entity"
)

type fakeAddressDB struct {
	execFn func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (f *fakeAddressDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if f.execFn != nil {
		return f.execFn(ctx, query, args...)
	}
	return fakeResult{lastInsertID: 1, rowsAffected: 1}, nil
}

func (f *fakeAddressDB) QueryRowContext(context.Context, string, ...interface{}) *sql.Row {
	return nil
}

func (f *fakeAddressDB) QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func TestAddressCreateSuccess(t *testing.T) {
	repo := NewAddressRepository(&fakeAddressDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{lastInsertID: 12}, nil
		},
	})

	address := &entity.Address{
		StreetName: "Street",
		StreenNo:   "10",
		City:       "City",
		County:     "County",
		Country:    "Country",
		ProfileID:  9,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := repo.Create(context.Background(), address); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if address.ID != 12 {
		t.Fatalf("expected ID 12, got %d", address.ID)
	}
}

func TestAddressCreateLastInsertIDError(t *testing.T) {
	repo := NewAddressRepository(&fakeAddressDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{lastInsertErr: errors.New("no id")}, nil
		},
	})

	err := repo.Create(context.Background(), &entity.Address{})
	if err == nil {
		t.Fatal("expected error when LastInsertId fails")
	}
}

func TestAddressUpdateNoopWhenNoRowsAffected(t *testing.T) {
	repo := NewAddressRepository(&fakeAddressDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{rowsAffected: 0}, nil
		},
	})

	if err := repo.Update(context.Background(), &entity.Address{ID: 1}); err != nil {
		t.Fatalf("expected no error for no-op update, got %v", err)
	}
}

func TestAddressDeleteNotFoundWhenNoRowsAffected(t *testing.T) {
	repo := NewAddressRepository(&fakeAddressDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{rowsAffected: 0}, nil
		},
	})

	err := repo.Delete(context.Background(), 1)
	if !errors.Is(err, ErrAddressNotFound) {
		t.Fatalf("expected ErrAddressNotFound, got %v", err)
	}
}
