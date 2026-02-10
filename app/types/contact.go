package types

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	contactDOBLayout      = "2006-01-02"
	maxContactPageSize    = 100
	defaultContactPage    = 1
	defaultContactPerPage = 20
)

type createContactBody struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Nin       string `json:"nin"`
	Dob       string `json:"dob"`
	Phone     string `json:"phone"`
	ProfileID uint64 `json:"profile_id"`
	Type      string `json:"type"`
}

type updateContactBody struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Nin       string `json:"nin"`
	Dob       string `json:"dob"`
	Phone     string `json:"phone"`
	ProfileID uint64 `json:"profile_id"`
	Type      string `json:"type"`
}

func NewCreateContactRequestFromContext(ctx echo.Context) (*CreateContactRequest, error) {
	var body createContactBody
	err := ctx.Bind(&body)
	if err != nil {
		return nil, err
	}

	return &CreateContactRequest{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Nin:       body.Nin,
		Dob:       body.Dob,
		Phone:     body.Phone,
		ProfileId: body.ProfileID,
		Type:      body.Type,
	}, nil
}

func (r *CreateContactRequest) Validate() error {
	rawDOB := strings.TrimSpace(r.Dob)
	if rawDOB != "" {
		if _, err := time.Parse(contactDOBLayout, rawDOB); err != nil {
			return errors.New("dob must be in YYYY-MM-DD format")
		}
	}
	if r.ProfileId == 0 {
		return errors.New("profile_id is required")
	}

	return nil
}

func NewGetContactRequestFromContext(ctx echo.Context) (*GetContactRequest, error) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	return &GetContactRequest{Id: id}, nil
}

func (r *GetContactRequest) Validate() error {
	if r.Id == 0 {
		return errors.New("invalid id provided")
	}

	return nil
}

func NewUpdateContactRequestFromContext(ctx echo.Context) (*UpdateContactRequest, error) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	var body updateContactBody
	err = ctx.Bind(&body)
	if err != nil {
		return nil, err
	}

	return &UpdateContactRequest{
		Id:        id,
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Nin:       body.Nin,
		Dob:       body.Dob,
		Phone:     body.Phone,
		ProfileId: body.ProfileID,
		Type:      body.Type,
	}, nil
}

func (r *UpdateContactRequest) Validate() error {
	if r.Id == 0 {
		return errors.New("invalid id provided")
	}
	rawDOB := strings.TrimSpace(r.Dob)
	if rawDOB != "" {
		if _, err := time.Parse(contactDOBLayout, rawDOB); err != nil {
			return errors.New("dob must be in YYYY-MM-DD format")
		}
	}
	if r.ProfileId == 0 {
		return errors.New("profile_id is required")
	}

	return nil
}

func NewDeleteContactRequestFromContext(ctx echo.Context) (*DeleteContactRequest, error) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	return &DeleteContactRequest{Id: id}, nil
}

func (r *DeleteContactRequest) Validate() error {
	if r.Id == 0 {
		return errors.New("invalid id provided")
	}

	return nil
}

func NewListContactsRequestFromContext(ctx echo.Context) (*ListContactsRequest, error) {
	req := &ListContactsRequest{
		Page:     defaultContactPage,
		PageSize: defaultContactPerPage,
	}

	if rawProfileID := strings.TrimSpace(ctx.QueryParam("profile_id")); rawProfileID != "" {
		profileID, err := strconv.ParseUint(rawProfileID, 10, 64)
		if err != nil {
			return nil, err
		}
		req.ProfileId = profileID
	}

	if rawPage := strings.TrimSpace(ctx.QueryParam("page")); rawPage != "" {
		page, err := strconv.ParseUint(rawPage, 10, 32)
		if err != nil {
			return nil, err
		}
		req.Page = uint32(page)
	}

	if rawPageSize := strings.TrimSpace(ctx.QueryParam("page_size")); rawPageSize != "" {
		pageSize, err := strconv.ParseUint(rawPageSize, 10, 32)
		if err != nil {
			return nil, err
		}
		req.PageSize = uint32(pageSize)
	}
	req.Type = strings.TrimSpace(ctx.QueryParam("type"))

	return req, nil
}

func (r *ListContactsRequest) Validate() error {
	if r.PageSize > maxContactPageSize {
		return errors.New("page_size must be less than or equal to 100")
	}

	return nil
}
