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
	ErrCompanyNotFound = errors.New("company not found")
)

type CompanyDBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type CompanyRepository struct {
	db CompanyDBTX
}

func NewCompanyRepository(db CompanyDBTX) *CompanyRepository {
	return &CompanyRepository{db: db}
}

func (r *CompanyRepository) Create(ctx context.Context, company *entity.Company) error {
	query := `
		INSERT INTO companies (name, registration_no, fiscal_code, profile_id, type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		company.Name,
		company.RegistrationNo,
		company.FiscalCode,
		company.ProfileID,
		company.Type,
		company.CreatedAt,
		company.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	company.ID = uint64(id)

	return nil
}

func (r *CompanyRepository) FindByID(ctx context.Context, id uint64) (*entity.Company, error) {
	query := `
		SELECT id, name, registration_no, fiscal_code, profile_id, type, created_at, updated_at
		FROM companies WHERE id = ?
	`
	company := &entity.Company{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&company.ID,
		&company.Name,
		&company.RegistrationNo,
		&company.FiscalCode,
		&company.ProfileID,
		&company.Type,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return company, nil
}

func (r *CompanyRepository) Update(ctx context.Context, company *entity.Company) error {
	query := `
		UPDATE companies SET
			name = ?,
			registration_no = ?,
			fiscal_code = ?,
			profile_id = ?,
			type = ?,
			updated_at = ?
		WHERE id = ?
	`
	company.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		company.Name,
		company.RegistrationNo,
		company.FiscalCode,
		company.ProfileID,
		company.Type,
		company.UpdatedAt,
		company.ID,
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

func (r *CompanyRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM companies WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrCompanyNotFound
	}

	return nil
}

func (r *CompanyRepository) List(ctx context.Context, profileID uint64, companyType string, limit, offset uint32) ([]*entity.Company, uint64, error) {
	if limit == 0 {
		limit = 20
	}

	companyType = strings.TrimSpace(companyType)
	whereClauses := make([]string, 0, 2)
	countArgs := make([]interface{}, 0, 2)
	if profileID > 0 {
		whereClauses = append(whereClauses, "profile_id = ?")
		countArgs = append(countArgs, profileID)
	}
	if companyType != "" {
		whereClauses = append(whereClauses, "`type` = ?")
		countArgs = append(countArgs, companyType)
	}

	countQuery := strings.Builder{}
	countQuery.WriteString(`SELECT COUNT(*) FROM companies`)
	if len(whereClauses) > 0 {
		countQuery.WriteString(` WHERE `)
		countQuery.WriteString(strings.Join(whereClauses, " AND "))
	}

	var total uint64
	if err := r.db.QueryRowContext(ctx, countQuery.String(), countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := strings.Builder{}
	query.WriteString(`
		SELECT id, name, registration_no, fiscal_code, profile_id, type, created_at, updated_at
		FROM companies
	`)
	args := make([]interface{}, 0, 4)
	if len(whereClauses) > 0 {
		query.WriteString(` WHERE `)
		query.WriteString(strings.Join(whereClauses, " AND "))
		args = append(args, countArgs...)
	}
	query.WriteString(` ORDER BY id DESC LIMIT ? OFFSET ?`)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	companies := make([]*entity.Company, 0)
	for rows.Next() {
		company := &entity.Company{}
		if err = rows.Scan(
			&company.ID,
			&company.Name,
			&company.RegistrationNo,
			&company.FiscalCode,
			&company.ProfileID,
			&company.Type,
			&company.CreatedAt,
			&company.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		companies = append(companies, company)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return companies, total, nil
}
