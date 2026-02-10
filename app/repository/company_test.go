package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/vibast-solutions/ms-go-profile/app/entity"
)

type fakeCompanyDB struct {
	execFn func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (f *fakeCompanyDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if f.execFn != nil {
		return f.execFn(ctx, query, args...)
	}
	return fakeResult{lastInsertID: 1, rowsAffected: 1}, nil
}

func (f *fakeCompanyDB) QueryRowContext(context.Context, string, ...interface{}) *sql.Row {
	return nil
}

func (f *fakeCompanyDB) QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func TestCompanyCreateSuccess(t *testing.T) {
	repo := NewCompanyRepository(&fakeCompanyDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{lastInsertID: 12}, nil
		},
	})

	company := &entity.Company{
		Name:           "ACME",
		RegistrationNo: "REG-1",
		FiscalCode:     "FISC-1",
		ProfileID:      9,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := repo.Create(context.Background(), company); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if company.ID != 12 {
		t.Fatalf("expected ID 12, got %d", company.ID)
	}
}

func TestCompanyCreateLastInsertIDError(t *testing.T) {
	repo := NewCompanyRepository(&fakeCompanyDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{lastInsertErr: errors.New("no id")}, nil
		},
	})

	err := repo.Create(context.Background(), &entity.Company{})
	if err == nil {
		t.Fatal("expected error when LastInsertId fails")
	}
}

func TestCompanyUpdateNoopWhenNoRowsAffected(t *testing.T) {
	repo := NewCompanyRepository(&fakeCompanyDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{rowsAffected: 0}, nil
		},
	})

	if err := repo.Update(context.Background(), &entity.Company{ID: 1}); err != nil {
		t.Fatalf("expected no error for no-op update, got %v", err)
	}
}

func TestCompanyDeleteNotFoundWhenNoRowsAffected(t *testing.T) {
	repo := NewCompanyRepository(&fakeCompanyDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{rowsAffected: 0}, nil
		},
	})

	err := repo.Delete(context.Background(), 1)
	if !errors.Is(err, ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
}
