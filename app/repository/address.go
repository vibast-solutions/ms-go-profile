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
	ErrAddressNotFound = errors.New("address not found")
)

type AddressDBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type AddressRepository struct {
	db AddressDBTX
}

func NewAddressRepository(db AddressDBTX) *AddressRepository {
	return &AddressRepository{db: db}
}

func (r *AddressRepository) Create(ctx context.Context, address *entity.Address) error {
	query := `
		INSERT INTO addresses (
			street_name, streen_no, city, county, country, profile_id,
			postal_code, building, apartment, additional_data, type,
			created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		address.StreetName,
		address.StreenNo,
		address.City,
		address.County,
		address.Country,
		address.ProfileID,
		address.PostalCode,
		address.Building,
		address.Apartment,
		address.AdditionalData,
		address.Type,
		address.CreatedAt,
		address.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	address.ID = uint64(id)

	return nil
}

func (r *AddressRepository) FindByID(ctx context.Context, id uint64) (*entity.Address, error) {
	query := `
		SELECT
			id, street_name, streen_no, city, county, country, profile_id,
			postal_code, building, apartment, additional_data, type,
			created_at, updated_at
		FROM addresses
		WHERE id = ?
	`
	address := &entity.Address{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&address.ID,
		&address.StreetName,
		&address.StreenNo,
		&address.City,
		&address.County,
		&address.Country,
		&address.ProfileID,
		&address.PostalCode,
		&address.Building,
		&address.Apartment,
		&address.AdditionalData,
		&address.Type,
		&address.CreatedAt,
		&address.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return address, nil
}

func (r *AddressRepository) Update(ctx context.Context, address *entity.Address) error {
	query := `
		UPDATE addresses SET
			street_name = ?,
			streen_no = ?,
			city = ?,
			county = ?,
			country = ?,
			profile_id = ?,
			postal_code = ?,
			building = ?,
			apartment = ?,
			additional_data = ?,
			type = ?,
			updated_at = ?
		WHERE id = ?
	`
	address.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		address.StreetName,
		address.StreenNo,
		address.City,
		address.County,
		address.Country,
		address.ProfileID,
		address.PostalCode,
		address.Building,
		address.Apartment,
		address.AdditionalData,
		address.Type,
		address.UpdatedAt,
		address.ID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return nil
	}

	return nil
}

func (r *AddressRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM addresses WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrAddressNotFound
	}

	return nil
}

func (r *AddressRepository) List(ctx context.Context, profileID uint64, limit, offset uint32) ([]*entity.Address, uint64, error) {
	if limit == 0 {
		limit = 20
	}

	countQuery := `SELECT COUNT(*) FROM addresses`
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
		SELECT
			id, street_name, streen_no, city, county, country, profile_id,
			postal_code, building, apartment, additional_data, type,
			created_at, updated_at
		FROM addresses
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

	addresses := make([]*entity.Address, 0)
	for rows.Next() {
		address := &entity.Address{}
		if err = rows.Scan(
			&address.ID,
			&address.StreetName,
			&address.StreenNo,
			&address.City,
			&address.County,
			&address.Country,
			&address.ProfileID,
			&address.PostalCode,
			&address.Building,
			&address.Apartment,
			&address.AdditionalData,
			&address.Type,
			&address.CreatedAt,
			&address.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		addresses = append(addresses, address)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return addresses, total, nil
}
