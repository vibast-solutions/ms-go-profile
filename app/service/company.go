package service

import (
	"context"
	"errors"
	"time"

	"github.com/vibast-solutions/ms-go-profile/app/entity"
	"github.com/vibast-solutions/ms-go-profile/app/repository"
)

var (
	ErrCompanyNotFound = errors.New("company not found")
)

type createCompanyRequest interface {
	GetName() string
	GetRegistrationNo() string
	GetFiscalCode() string
	GetProfileId() uint64
	GetType() string
}

type updateCompanyRequest interface {
	GetId() uint64
	GetName() string
	GetRegistrationNo() string
	GetFiscalCode() string
	GetProfileId() uint64
	GetType() string
}

type listCompaniesRequest interface {
	GetProfileId() uint64
	GetPage() uint32
	GetPageSize() uint32
	GetType() string
}

type companyRepository interface {
	Create(ctx context.Context, company *entity.Company) error
	FindByID(ctx context.Context, id uint64) (*entity.Company, error)
	Update(ctx context.Context, company *entity.Company) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, profileID uint64, companyType string, limit, offset uint32) ([]*entity.Company, uint64, error)
}

type CompanyList struct {
	Companies []*entity.Company
	Page      uint32
	PageSize  uint32
	Total     uint64
}

type CompanyService struct {
	companyRepo companyRepository
}

func NewCompanyService(companyRepo companyRepository) *CompanyService {
	return &CompanyService{companyRepo: companyRepo}
}

func (s *CompanyService) Create(ctx context.Context, req createCompanyRequest) (*entity.Company, error) {
	now := time.Now()
	company := &entity.Company{
		Name:           req.GetName(),
		RegistrationNo: req.GetRegistrationNo(),
		FiscalCode:     req.GetFiscalCode(),
		ProfileID:      req.GetProfileId(),
		Type:           req.GetType(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.companyRepo.Create(ctx, company); err != nil {
		return nil, err
	}

	return company, nil
}

func (s *CompanyService) GetByID(ctx context.Context, id uint64) (*entity.Company, error) {
	company, err := s.companyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, ErrCompanyNotFound
	}

	return company, nil
}

func (s *CompanyService) Update(ctx context.Context, req updateCompanyRequest) (*entity.Company, error) {
	company, err := s.companyRepo.FindByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, ErrCompanyNotFound
	}

	company.Name = req.GetName()
	company.RegistrationNo = req.GetRegistrationNo()
	company.FiscalCode = req.GetFiscalCode()
	company.ProfileID = req.GetProfileId()
	company.Type = req.GetType()

	if err = s.companyRepo.Update(ctx, company); err != nil {
		if errors.Is(err, repository.ErrCompanyNotFound) {
			return nil, ErrCompanyNotFound
		}
		return nil, err
	}

	return company, nil
}

func (s *CompanyService) Delete(ctx context.Context, id uint64) error {
	if err := s.companyRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrCompanyNotFound) {
			return ErrCompanyNotFound
		}
		return err
	}

	return nil
}

func (s *CompanyService) List(ctx context.Context, req listCompaniesRequest) (*CompanyList, error) {
	page := req.GetPage()
	if page == 0 {
		page = 1
	}

	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	companies, total, err := s.companyRepo.List(ctx, req.GetProfileId(), req.GetType(), pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &CompanyList{
		Companies: companies,
		Page:      page,
		PageSize:  pageSize,
		Total:     total,
	}, nil
}
