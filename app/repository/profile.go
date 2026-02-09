package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/vibast-solutions/ms-go-profile/app/entity"
)

var (
	ErrProfileNotFound      = errors.New("profile not found")
	ErrProfileAlreadyExists = errors.New("profile already exists")
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type ProfileRepository struct {
	db DBTX
}

func NewProfileRepository(db DBTX) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) WithTx(tx *sql.Tx) *ProfileRepository {
	return &ProfileRepository{db: tx}
}

func (r *ProfileRepository) Create(ctx context.Context, profile *entity.Profile) error {
	query := `
		INSERT INTO profile (user_id, email, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		profile.UserID,
		profile.Email,
		profile.CreatedAt,
		profile.UpdatedAt,
	)
	if err != nil {
		if isDuplicateEntryError(err) {
			return ErrProfileAlreadyExists
		}
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	profile.ID = uint64(id)
	return nil
}

func (r *ProfileRepository) FindByID(ctx context.Context, id uint64) (*entity.Profile, error) {
	query := `
		SELECT id, user_id, email, created_at, updated_at
		FROM profile WHERE id = ?
	`
	profile := &entity.Profile{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Email,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (r *ProfileRepository) FindByUserID(ctx context.Context, userID uint64) (*entity.Profile, error) {
	query := `
		SELECT id, user_id, email, created_at, updated_at
		FROM profile WHERE user_id = ?
	`
	profile := &entity.Profile{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Email,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (r *ProfileRepository) Update(ctx context.Context, profile *entity.Profile) error {
	query := `
		UPDATE profile SET
			email = ?,
			updated_at = ?
		WHERE id = ?
	`
	profile.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		profile.Email,
		profile.UpdatedAt,
		profile.ID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrProfileNotFound
	}

	return nil
}

func (r *ProfileRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM profile WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrProfileNotFound
	}

	return nil
}

func isDuplicateEntryError(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
