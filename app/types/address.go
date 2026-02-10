package types

import (
	"errors"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	maxAddressPageSize    = 100
	defaultAddressPage    = 1
	defaultAddressPerPage = 20
)

type createAddressBody struct {
	StreetName     string `json:"street_name"`
	StreenNo       string `json:"streen_no"`
	City           string `json:"city"`
	County         string `json:"county"`
	Country        string `json:"country"`
	ProfileID      uint64 `json:"profile_id"`
	PostalCode     string `json:"postal_code"`
	Building       string `json:"building"`
	Apartment      string `json:"apartment"`
	AdditionalData string `json:"additional_data"`
	Type           string `json:"type"`
}

type updateAddressBody struct {
	ID             uint64 `param:"id"`
	StreetName     string `json:"street_name"`
	StreenNo       string `json:"streen_no"`
	City           string `json:"city"`
	County         string `json:"county"`
	Country        string `json:"country"`
	ProfileID      uint64 `json:"profile_id"`
	PostalCode     string `json:"postal_code"`
	Building       string `json:"building"`
	Apartment      string `json:"apartment"`
	AdditionalData string `json:"additional_data"`
	Type           string `json:"type"`
}

type addressPathParams struct {
	ID uint64 `param:"id"`
}

type listAddressesQuery struct {
	ProfileID uint64 `query:"profile_id"`
	Page      uint32 `query:"page"`
	PageSize  uint32 `query:"page_size"`
	Type      string `query:"type"`
}

func NewCreateAddressRequestFromContext(ctx echo.Context) (*CreateAddressRequest, error) {
	var body createAddressBody
	if err := ctx.Bind(&body); err != nil {
		return nil, err
	}

	return &CreateAddressRequest{
		StreetName:     body.StreetName,
		StreenNo:       body.StreenNo,
		City:           body.City,
		County:         body.County,
		Country:        body.Country,
		ProfileId:      body.ProfileID,
		PostalCode:     body.PostalCode,
		Building:       body.Building,
		Apartment:      body.Apartment,
		AdditionalData: body.AdditionalData,
		Type:           body.Type,
	}, nil
}

func (r *CreateAddressRequest) Validate() error {
	if strings.TrimSpace(r.StreetName) == "" {
		return errors.New("street_name is required")
	}
	if strings.TrimSpace(r.StreenNo) == "" {
		return errors.New("streen_no is required")
	}
	if strings.TrimSpace(r.City) == "" {
		return errors.New("city is required")
	}
	if strings.TrimSpace(r.County) == "" {
		return errors.New("county is required")
	}
	if strings.TrimSpace(r.Country) == "" {
		return errors.New("country is required")
	}
	if r.ProfileId == 0 {
		return errors.New("profile_id is required")
	}
	if len(r.AdditionalData) > 512 {
		return errors.New("additional_data must be less than or equal to 512 characters")
	}

	return nil
}

func NewGetAddressRequestFromContext(ctx echo.Context) (*GetAddressRequest, error) {
	params := &addressPathParams{}
	if err := ctx.Bind(params); err != nil {
		return nil, err
	}

	return &GetAddressRequest{Id: params.ID}, nil
}

func (r *GetAddressRequest) Validate() error {
	if r.Id == 0 {
		return errors.New("invalid id provided")
	}

	return nil
}

func NewUpdateAddressRequestFromContext(ctx echo.Context) (*UpdateAddressRequest, error) {
	var body updateAddressBody
	if err := ctx.Bind(&body); err != nil {
		return nil, err
	}

	return &UpdateAddressRequest{
		Id:             body.ID,
		StreetName:     body.StreetName,
		StreenNo:       body.StreenNo,
		City:           body.City,
		County:         body.County,
		Country:        body.Country,
		ProfileId:      body.ProfileID,
		PostalCode:     body.PostalCode,
		Building:       body.Building,
		Apartment:      body.Apartment,
		AdditionalData: body.AdditionalData,
		Type:           body.Type,
	}, nil
}

func (r *UpdateAddressRequest) Validate() error {
	if r.Id == 0 {
		return errors.New("invalid id provided")
	}
	if strings.TrimSpace(r.StreetName) == "" {
		return errors.New("street_name is required")
	}
	if strings.TrimSpace(r.StreenNo) == "" {
		return errors.New("streen_no is required")
	}
	if strings.TrimSpace(r.City) == "" {
		return errors.New("city is required")
	}
	if strings.TrimSpace(r.County) == "" {
		return errors.New("county is required")
	}
	if strings.TrimSpace(r.Country) == "" {
		return errors.New("country is required")
	}
	if r.ProfileId == 0 {
		return errors.New("profile_id is required")
	}
	if len(r.AdditionalData) > 512 {
		return errors.New("additional_data must be less than or equal to 512 characters")
	}

	return nil
}

func NewDeleteAddressRequestFromContext(ctx echo.Context) (*DeleteAddressRequest, error) {
	params := &addressPathParams{}
	if err := ctx.Bind(params); err != nil {
		return nil, err
	}

	return &DeleteAddressRequest{Id: params.ID}, nil
}

func (r *DeleteAddressRequest) Validate() error {
	if r.Id == 0 {
		return errors.New("invalid id provided")
	}

	return nil
}

func NewListAddressesRequestFromContext(ctx echo.Context) (*ListAddressesRequest, error) {
	query := &listAddressesQuery{
		Page:     defaultAddressPage,
		PageSize: defaultAddressPerPage,
	}
	if err := ctx.Bind(query); err != nil {
		return nil, err
	}

	return &ListAddressesRequest{
		ProfileId: query.ProfileID,
		Page:      query.Page,
		PageSize:  query.PageSize,
		Type:      strings.TrimSpace(query.Type),
	}, nil
}

func (r *ListAddressesRequest) Validate() error {
	if r.ProfileId == 0 {
		return errors.New("profile_id is required")
	}
	if r.PageSize > maxAddressPageSize {
		return errors.New("page_size must be less than or equal to 100")
	}

	return nil
}
