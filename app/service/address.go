package service

import (
	"context"
	"errors"
	"time"

	"github.com/vibast-solutions/ms-go-profile/app/entity"
	"github.com/vibast-solutions/ms-go-profile/app/repository"
)

var (
	ErrAddressNotFound = errors.New("address not found")
)

type createAddressRequest interface {
	GetStreetName() string
	GetStreenNo() string
	GetCity() string
	GetCounty() string
	GetCountry() string
	GetProfileId() uint64
	GetPostalCode() string
	GetBuilding() string
	GetApartment() string
	GetAdditionalData() string
	GetType() string
}

type updateAddressRequest interface {
	GetId() uint64
	GetStreetName() string
	GetStreenNo() string
	GetCity() string
	GetCounty() string
	GetCountry() string
	GetProfileId() uint64
	GetPostalCode() string
	GetBuilding() string
	GetApartment() string
	GetAdditionalData() string
	GetType() string
}

type listAddressesRequest interface {
	GetProfileId() uint64
	GetPage() uint32
	GetPageSize() uint32
}

type addressRepository interface {
	Create(ctx context.Context, address *entity.Address) error
	FindByID(ctx context.Context, id uint64) (*entity.Address, error)
	Update(ctx context.Context, address *entity.Address) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, profileID uint64, limit, offset uint32) ([]*entity.Address, uint64, error)
}

type AddressList struct {
	Addresses []*entity.Address
	Page      uint32
	PageSize  uint32
	Total     uint64
}

type AddressService struct {
	addressRepo addressRepository
}

func NewAddressService(addressRepo addressRepository) *AddressService {
	return &AddressService{addressRepo: addressRepo}
}

func (s *AddressService) Create(ctx context.Context, req createAddressRequest) (*entity.Address, error) {
	now := time.Now()
	address := &entity.Address{
		StreetName:     req.GetStreetName(),
		StreenNo:       req.GetStreenNo(),
		City:           req.GetCity(),
		County:         req.GetCounty(),
		Country:        req.GetCountry(),
		ProfileID:      req.GetProfileId(),
		PostalCode:     req.GetPostalCode(),
		Building:       req.GetBuilding(),
		Apartment:      req.GetApartment(),
		AdditionalData: req.GetAdditionalData(),
		Type:           req.GetType(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.addressRepo.Create(ctx, address); err != nil {
		return nil, err
	}

	return address, nil
}

func (s *AddressService) GetByID(ctx context.Context, id uint64) (*entity.Address, error) {
	address, err := s.addressRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if address == nil {
		return nil, ErrAddressNotFound
	}

	return address, nil
}

func (s *AddressService) Update(ctx context.Context, req updateAddressRequest) (*entity.Address, error) {
	address, err := s.addressRepo.FindByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if address == nil {
		return nil, ErrAddressNotFound
	}

	address.StreetName = req.GetStreetName()
	address.StreenNo = req.GetStreenNo()
	address.City = req.GetCity()
	address.County = req.GetCounty()
	address.Country = req.GetCountry()
	address.ProfileID = req.GetProfileId()
	address.PostalCode = req.GetPostalCode()
	address.Building = req.GetBuilding()
	address.Apartment = req.GetApartment()
	address.AdditionalData = req.GetAdditionalData()
	address.Type = req.GetType()

	if err = s.addressRepo.Update(ctx, address); err != nil {
		if errors.Is(err, repository.ErrAddressNotFound) {
			return nil, ErrAddressNotFound
		}
		return nil, err
	}

	return address, nil
}

func (s *AddressService) Delete(ctx context.Context, id uint64) error {
	if err := s.addressRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrAddressNotFound) {
			return ErrAddressNotFound
		}
		return err
	}

	return nil
}

func (s *AddressService) List(ctx context.Context, req listAddressesRequest) (*AddressList, error) {
	page := req.GetPage()
	if page == 0 {
		page = 1
	}

	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	addresses, total, err := s.addressRepo.List(ctx, req.GetProfileId(), pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &AddressList{
		Addresses: addresses,
		Page:      page,
		PageSize:  pageSize,
		Total:     total,
	}, nil
}
