package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/vibast-solutions/ms-go-profile/app/entity"
)

type fakeContactDB struct {
	execFn func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (f *fakeContactDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if f.execFn != nil {
		return f.execFn(ctx, query, args...)
	}
	return fakeResult{lastInsertID: 1, rowsAffected: 1}, nil
}

func (f *fakeContactDB) QueryRowContext(context.Context, string, ...interface{}) *sql.Row {
	// Not used in these tests.
	return nil
}

func (f *fakeContactDB) QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error) {
	// Not used in these tests.
	return nil, nil
}

func TestContactCreateSuccess(t *testing.T) {
	dob := time.Now()
	repo := NewContactRepository(&fakeContactDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{lastInsertID: 9}, nil
		},
	})

	contact := &entity.Contact{
		FirstName: "John",
		LastName:  "Doe",
		NIN:       "1234",
		DOB:       &dob,
		Phone:     "123456",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ProfileID: 5,
	}
	if err := repo.Create(context.Background(), contact); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if contact.ID != 9 {
		t.Fatalf("expected ID 9, got %d", contact.ID)
	}
}

func TestContactCreateLastInsertIDError(t *testing.T) {
	repo := NewContactRepository(&fakeContactDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{lastInsertErr: errors.New("no id")}, nil
		},
	})

	err := repo.Create(context.Background(), &entity.Contact{})
	if err == nil {
		t.Fatal("expected error when LastInsertId fails")
	}
}

func TestContactUpdateNoopWhenNoRowsAffected(t *testing.T) {
	repo := NewContactRepository(&fakeContactDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{rowsAffected: 0}, nil
		},
	})

	err := repo.Update(context.Background(), &entity.Contact{ID: 1})
	if err != nil {
		t.Fatalf("expected no error for no-op update, got %v", err)
	}
}

func TestContactDeleteNotFoundWhenNoRowsAffected(t *testing.T) {
	repo := NewContactRepository(&fakeContactDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{rowsAffected: 0}, nil
		},
	})

	err := repo.Delete(context.Background(), 1)
	if !errors.Is(err, ErrContactNotFound) {
		t.Fatalf("expected ErrContactNotFound, got %v", err)
	}
}
