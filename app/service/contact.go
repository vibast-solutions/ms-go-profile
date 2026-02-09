package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/vibast-solutions/ms-go-profile/app/entity"
	"github.com/vibast-solutions/ms-go-profile/app/repository"
)

var (
	ErrContactNotFound = errors.New("contact not found")
)

const contactDOBLayout = "2006-01-02"

type createContactRequest interface {
	GetFirstName() string
	GetLastName() string
	GetNin() string
	GetDob() string
	GetPhone() string
	GetProfileId() uint64
	GetType() string
}

type updateContactRequest interface {
	GetId() uint64
	GetFirstName() string
	GetLastName() string
	GetNin() string
	GetDob() string
	GetPhone() string
	GetProfileId() uint64
	GetType() string
}

type listContactsRequest interface {
	GetProfileId() uint64
	GetPage() uint32
	GetPageSize() uint32
}

type contactRepository interface {
	Create(ctx context.Context, contact *entity.Contact) error
	FindByID(ctx context.Context, id uint64) (*entity.Contact, error)
	Update(ctx context.Context, contact *entity.Contact) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, profileID uint64, limit, offset uint32) ([]*entity.Contact, uint64, error)
}

type ContactList struct {
	Contacts []*entity.Contact
	Page     uint32
	PageSize uint32
	Total    uint64
}

type ContactService struct {
	contactRepo contactRepository
}

func NewContactService(contactRepo contactRepository) *ContactService {
	return &ContactService{contactRepo: contactRepo}
}

func (s *ContactService) Create(ctx context.Context, req createContactRequest) (*entity.Contact, error) {
	dob, err := parseOptionalContactDOB(req.GetDob())
	if err != nil {
		return nil, err
	}

	now := time.Now()
	contact := &entity.Contact{
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		NIN:       req.GetNin(),
		DOB:       dob,
		Phone:     req.GetPhone(),
		Type:      req.GetType(),
		CreatedAt: now,
		UpdatedAt: now,
		ProfileID: req.GetProfileId(),
	}

	if err = s.contactRepo.Create(ctx, contact); err != nil {
		return nil, err
	}

	return contact, nil
}

func (s *ContactService) GetByID(ctx context.Context, id uint64) (*entity.Contact, error) {
	contact, err := s.contactRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if contact == nil {
		return nil, ErrContactNotFound
	}

	return contact, nil
}

func (s *ContactService) Update(ctx context.Context, req updateContactRequest) (*entity.Contact, error) {
	contact, err := s.contactRepo.FindByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if contact == nil {
		return nil, ErrContactNotFound
	}

	dob, err := parseOptionalContactDOB(req.GetDob())
	if err != nil {
		return nil, err
	}

	contact.FirstName = req.GetFirstName()
	contact.LastName = req.GetLastName()
	contact.NIN = req.GetNin()
	contact.DOB = dob
	contact.Phone = req.GetPhone()
	contact.ProfileID = req.GetProfileId()
	contact.Type = req.GetType()

	if err = s.contactRepo.Update(ctx, contact); err != nil {
		if errors.Is(err, repository.ErrContactNotFound) {
			return nil, ErrContactNotFound
		}
		return nil, err
	}

	return contact, nil
}

func (s *ContactService) Delete(ctx context.Context, id uint64) error {
	if err := s.contactRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrContactNotFound) {
			return ErrContactNotFound
		}
		return err
	}

	return nil
}

func (s *ContactService) List(ctx context.Context, req listContactsRequest) (*ContactList, error) {
	page := req.GetPage()
	if page == 0 {
		page = 1
	}

	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	contacts, total, err := s.contactRepo.List(ctx, req.GetProfileId(), pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &ContactList{
		Contacts: contacts,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func parseOptionalContactDOB(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	dob, err := time.Parse(contactDOBLayout, raw)
	if err != nil {
		return nil, err
	}

	return &dob, nil
}
