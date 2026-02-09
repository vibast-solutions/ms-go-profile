package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/vibast-solutions/ms-go-profile/app/entity"
)

var (
	ErrContactNotFound = errors.New("contact not found")
)

type ContactDBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type ContactRepository struct {
	db ContactDBTX
}

func NewContactRepository(db ContactDBTX) *ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) Create(ctx context.Context, contact *entity.Contact) error {
	query := `
		INSERT INTO contacts (first_name, last_name, nin, dob, phone, created_at, updated_at, profile_id, type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		contact.FirstName,
		contact.LastName,
		contact.NIN,
		contact.DOB,
		contact.Phone,
		contact.CreatedAt,
		contact.UpdatedAt,
		contact.ProfileID,
		contact.Type,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	contact.ID = uint64(id)

	return nil
}

func (r *ContactRepository) FindByID(ctx context.Context, id uint64) (*entity.Contact, error) {
	query := `
		SELECT id, first_name, last_name, nin, dob, phone, created_at, updated_at, profile_id, type
		FROM contacts WHERE id = ?
	`
	contact := &entity.Contact{}
	var dob sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&contact.ID,
		&contact.FirstName,
		&contact.LastName,
		&contact.NIN,
		&dob,
		&contact.Phone,
		&contact.CreatedAt,
		&contact.UpdatedAt,
		&contact.ProfileID,
		&contact.Type,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if dob.Valid {
		contact.DOB = &dob.Time
	}

	return contact, nil
}

func (r *ContactRepository) Update(ctx context.Context, contact *entity.Contact) error {
	query := `
		UPDATE contacts SET
			first_name = ?,
			last_name = ?,
			nin = ?,
			dob = ?,
			phone = ?,
			updated_at = ?,
			profile_id = ?,
			type = ?
		WHERE id = ?
	`
	contact.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		contact.FirstName,
		contact.LastName,
		contact.NIN,
		contact.DOB,
		contact.Phone,
		contact.UpdatedAt,
		contact.ProfileID,
		contact.Type,
		contact.ID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		// MySQL reports 0 affected rows when values are unchanged.
		// The service already checks existence before update, so treat this as success.
		return nil
	}

	return nil
}

func (r *ContactRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM contacts WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrContactNotFound
	}

	return nil
}

func (r *ContactRepository) List(ctx context.Context, profileID uint64, limit, offset uint32) ([]*entity.Contact, uint64, error) {
	if limit == 0 {
		limit = 20
	}

	countQuery := `SELECT COUNT(*) FROM contacts`
	countArgs := make([]interface{}, 0, 1)
	if profileID > 0 {
		countQuery += ` WHERE profile_id = ?`
		countArgs = append(countArgs, profileID)
	}

	var total uint64
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := strings.Builder{}
	query.WriteString(`
		SELECT id, first_name, last_name, nin, dob, phone, created_at, updated_at, profile_id, type
		FROM contacts
	`)
	args := make([]interface{}, 0, 3)
	if profileID > 0 {
		query.WriteString(` WHERE profile_id = ?`)
		args = append(args, profileID)
	}
	query.WriteString(` ORDER BY id DESC LIMIT ? OFFSET ?`)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	contacts := make([]*entity.Contact, 0)
	for rows.Next() {
		contact := &entity.Contact{}
		var dob sql.NullTime
		if err = rows.Scan(
			&contact.ID,
			&contact.FirstName,
			&contact.LastName,
			&contact.NIN,
			&dob,
			&contact.Phone,
			&contact.CreatedAt,
			&contact.UpdatedAt,
			&contact.ProfileID,
			&contact.Type,
		); err != nil {
			return nil, 0, err
		}
		if dob.Valid {
			contact.DOB = &dob.Time
		}
		contacts = append(contacts, contact)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return contacts, total, nil
}
