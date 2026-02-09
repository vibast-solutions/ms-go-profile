package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/vibast-solutions/ms-go-profile/app/entity"
)

type fakeDB struct {
	execFn func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (f *fakeDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if f.execFn != nil {
		return f.execFn(ctx, query, args...)
	}
	return fakeResult{lastInsertID: 1, rowsAffected: 1}, nil
}

func (f *fakeDB) QueryRowContext(context.Context, string, ...interface{}) *sql.Row {
	// Not used in these tests.
	return nil
}

type fakeResult struct {
	lastInsertID    int64
	rowsAffected    int64
	lastInsertErr   error
	rowsAffectedErr error
}

func (r fakeResult) LastInsertId() (int64, error) {
	return r.lastInsertID, r.lastInsertErr
}

func (r fakeResult) RowsAffected() (int64, error) {
	return r.rowsAffected, r.rowsAffectedErr
}

func TestCreateSuccess(t *testing.T) {
	repo := NewProfileRepository(&fakeDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{lastInsertID: 13}, nil
		},
	})
	profile := &entity.Profile{
		UserID:    42,
		Email:     "john@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), profile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if profile.ID != 13 {
		t.Fatalf("expected ID to be set to 13, got %d", profile.ID)
	}
}

func TestCreateMapsDuplicateError(t *testing.T) {
	dupErr := &mysqlDriver.MySQLError{Number: 1062, Message: "Duplicate entry"}
	repo := NewProfileRepository(&fakeDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return nil, dupErr
		},
	})

	err := repo.Create(context.Background(), &entity.Profile{})
	if !errors.Is(err, ErrProfileAlreadyExists) {
		t.Fatalf("expected ErrProfileAlreadyExists, got: %v", err)
	}
}

func TestCreateLastInsertIDError(t *testing.T) {
	repo := NewProfileRepository(&fakeDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{lastInsertErr: errors.New("no last insert id")}, nil
		},
	})

	err := repo.Create(context.Background(), &entity.Profile{})
	if err == nil {
		t.Fatal("expected error when LastInsertId fails")
	}
}

func TestUpdateSuccess(t *testing.T) {
	profile := &entity.Profile{ID: 11, Email: "before@example.com", UpdatedAt: time.Now().Add(-time.Hour)}
	oldUpdatedAt := profile.UpdatedAt

	repo := NewProfileRepository(&fakeDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{rowsAffected: 1}, nil
		},
	})

	if err := repo.Update(context.Background(), profile); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !profile.UpdatedAt.After(oldUpdatedAt) {
		t.Fatalf("expected UpdatedAt to move forward (old=%v new=%v)", oldUpdatedAt, profile.UpdatedAt)
	}
}

func TestUpdateNotFoundWhenNoRowsAffected(t *testing.T) {
	repo := NewProfileRepository(&fakeDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{rowsAffected: 0}, nil
		},
	})

	err := repo.Update(context.Background(), &entity.Profile{ID: 11})
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound, got: %v", err)
	}
}

func TestDeleteSuccess(t *testing.T) {
	repo := NewProfileRepository(&fakeDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{rowsAffected: 1}, nil
		},
	})

	if err := repo.Delete(context.Background(), 10); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestDeleteNotFoundWhenNoRowsAffected(t *testing.T) {
	repo := NewProfileRepository(&fakeDB{
		execFn: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return fakeResult{rowsAffected: 0}, nil
		},
	})

	err := repo.Delete(context.Background(), 10)
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound, got: %v", err)
	}
}

func TestIsDuplicateEntryError(t *testing.T) {
	if !isDuplicateEntryError(&mysqlDriver.MySQLError{Number: 1062}) {
		t.Fatal("expected duplicate-entry detection to be true for MySQL 1062")
	}
	if isDuplicateEntryError(errors.New("boom")) {
		t.Fatal("expected duplicate-entry detection to be false for generic error")
	}
}

var (
	queryDriverOnce sync.Once
	queryCaseID     uint64
	queryCases      sync.Map
)

type queryCase struct {
	queryErr error
	row      []driver.Value
}

type queryStubDriver struct{}

func (d *queryStubDriver) Open(name string) (driver.Conn, error) {
	return &queryStubConn{dsn: name}, nil
}

type queryStubConn struct {
	dsn string
}

func (c *queryStubConn) Prepare(query string) (driver.Stmt, error) {
	return &queryStubStmt{query: query}, nil
}

func (c *queryStubConn) Close() error { return nil }

func (c *queryStubConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are not supported in query stub")
}

func (c *queryStubConn) QueryContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
	v, ok := queryCases.Load(c.dsn)
	if !ok {
		return nil, errors.New("missing query case")
	}
	tc := v.(queryCase)
	if tc.queryErr != nil {
		return nil, tc.queryErr
	}

	return &queryStubRows{
		columns:  []string{"id", "user_id", "email", "created_at", "updated_at"},
		row:      tc.row,
		returned: false,
	}, nil
}

type queryStubStmt struct {
	query string
}

func (s *queryStubStmt) Close() error { return nil }

func (s *queryStubStmt) NumInput() int { return -1 }

func (s *queryStubStmt) Exec(_ []driver.Value) (driver.Result, error) {
	return nil, errors.New("exec is not supported in query stub")
}

func (s *queryStubStmt) Query(_ []driver.Value) (driver.Rows, error) {
	return nil, errors.New("query is not supported in prepared stub")
}

type queryStubRows struct {
	columns  []string
	row      []driver.Value
	returned bool
}

func (r *queryStubRows) Columns() []string { return r.columns }

func (r *queryStubRows) Close() error { return nil }

func (r *queryStubRows) Next(dest []driver.Value) error {
	if r.returned || r.row == nil {
		return io.EOF
	}
	copy(dest, r.row)
	r.returned = true
	return nil
}

func newQueryTestDB(t *testing.T, tc queryCase) *sql.DB {
	t.Helper()

	queryDriverOnce.Do(func() {
		sql.Register("repo_query_stub", &queryStubDriver{})
	})

	id := atomic.AddUint64(&queryCaseID, 1)
	dsn := fmt.Sprintf("case-%d", id)
	queryCases.Store(dsn, tc)

	db, err := sql.Open("repo_query_stub", dsn)
	if err != nil {
		t.Fatalf("failed to open query stub db: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
		queryCases.Delete(dsn)
	})

	return db
}

func TestFindByIDNoRows(t *testing.T) {
	db := newQueryTestDB(t, queryCase{row: nil})
	repo := NewProfileRepository(db)

	profile, err := repo.FindByID(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if profile != nil {
		t.Fatalf("expected nil profile for no rows, got: %+v", profile)
	}
}

func TestFindByIDSuccess(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	db := newQueryTestDB(t, queryCase{
		row: []driver.Value{int64(3), int64(42), "john@example.com", now, now},
	})
	repo := NewProfileRepository(db)

	profile, err := repo.FindByID(context.Background(), 3)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if profile == nil {
		t.Fatal("expected profile, got nil")
	}
	if profile.ID != 3 || profile.UserID != 42 || profile.Email != "john@example.com" {
		t.Fatalf("unexpected profile values: %+v", profile)
	}
}

func TestFindByIDQueryError(t *testing.T) {
	db := newQueryTestDB(t, queryCase{
		queryErr: errors.New("query failed"),
	})
	repo := NewProfileRepository(db)

	_, err := repo.FindByID(context.Background(), 3)
	if err == nil || !strings.Contains(err.Error(), "query failed") {
		t.Fatalf("expected query error, got: %v", err)
	}
}

func TestFindByIDScanError(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	db := newQueryTestDB(t, queryCase{
		row: []driver.Value{"bad-id", int64(42), "john@example.com", now, now},
	})
	repo := NewProfileRepository(db)

	_, err := repo.FindByID(context.Background(), 3)
	if err == nil {
		t.Fatal("expected scan error, got nil")
	}
}

func TestFindByUserIDNoRows(t *testing.T) {
	db := newQueryTestDB(t, queryCase{row: nil})
	repo := NewProfileRepository(db)

	profile, err := repo.FindByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if profile != nil {
		t.Fatalf("expected nil profile for no rows, got: %+v", profile)
	}
}

func TestFindByUserIDSuccess(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	db := newQueryTestDB(t, queryCase{
		row: []driver.Value{int64(3), int64(42), "john@example.com", now, now},
	})
	repo := NewProfileRepository(db)

	profile, err := repo.FindByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if profile == nil {
		t.Fatal("expected profile, got nil")
	}
	if profile.ID != 3 || profile.UserID != 42 || profile.Email != "john@example.com" {
		t.Fatalf("unexpected profile values: %+v", profile)
	}
}
