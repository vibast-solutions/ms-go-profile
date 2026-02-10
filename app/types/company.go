package types

import (
	"errors"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	maxCompanyPageSize    = 100
	defaultCompanyPage    = 1
	defaultCompanyPerPage = 20
)

type createCompanyBody struct {
	Name           string `json:"name"`
	RegistrationNo string `json:"registration_no"`
	FiscalCode     string `json:"fiscal_code"`
	ProfileID      uint64 `json:"profile_id"`
	Type           string `json:"type"`
}

type updateCompanyBody struct {
	Name           string `json:"name"`
	RegistrationNo string `json:"registration_no"`
	FiscalCode     string `json:"fiscal_code"`
	ProfileID      uint64 `json:"profile_id"`
	Type           string `json:"type"`
}

func NewCreateCompanyRequestFromContext(ctx echo.Context) (*CreateCompanyRequest, error) {
	var body createCompanyBody
	if err := ctx.Bind(&body); err != nil {
		return nil, err
	}

	return &CreateCompanyRequest{
		Name:           body.Name,
		RegistrationNo: body.RegistrationNo,
		FiscalCode:     body.FiscalCode,
		ProfileId:      body.ProfileID,
		Type:           body.Type,
	}, nil
}

func (r *CreateCompanyRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(r.RegistrationNo) == "" {
		return errors.New("registration_no is required")
	}
	if strings.TrimSpace(r.FiscalCode) == "" {
		return errors.New("fiscal_code is required")
	}
	if r.ProfileId == 0 {
		return errors.New("profile_id is required")
	}

	return nil
}

func NewGetCompanyRequestFromContext(ctx echo.Context) (*GetCompanyRequest, error) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	return &GetCompanyRequest{Id: id}, nil
}

func (r *GetCompanyRequest) Validate() error {
	if r.Id == 0 {
		return errors.New("invalid id provided")
	}

	return nil
}

func NewUpdateCompanyRequestFromContext(ctx echo.Context) (*UpdateCompanyRequest, error) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	var body updateCompanyBody
	if err = ctx.Bind(&body); err != nil {
		return nil, err
	}

	return &UpdateCompanyRequest{
		Id:             id,
		Name:           body.Name,
		RegistrationNo: body.RegistrationNo,
		FiscalCode:     body.FiscalCode,
		ProfileId:      body.ProfileID,
		Type:           body.Type,
	}, nil
}

func (r *UpdateCompanyRequest) Validate() error {
	if r.Id == 0 {
		return errors.New("invalid id provided")
	}
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(r.RegistrationNo) == "" {
		return errors.New("registration_no is required")
	}
	if strings.TrimSpace(r.FiscalCode) == "" {
		return errors.New("fiscal_code is required")
	}
	if r.ProfileId == 0 {
		return errors.New("profile_id is required")
	}

	return nil
}

func NewDeleteCompanyRequestFromContext(ctx echo.Context) (*DeleteCompanyRequest, error) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	return &DeleteCompanyRequest{Id: id}, nil
}

func (r *DeleteCompanyRequest) Validate() error {
	if r.Id == 0 {
		return errors.New("invalid id provided")
	}

	return nil
}

func NewListCompaniesRequestFromContext(ctx echo.Context) (*ListCompaniesRequest, error) {
	req := &ListCompaniesRequest{
		Page:     defaultCompanyPage,
		PageSize: defaultCompanyPerPage,
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

func (r *ListCompaniesRequest) Validate() error {
	if r.ProfileId == 0 {
		return errors.New("profile_id is required")
	}
	if r.PageSize > maxCompanyPageSize {
		return errors.New("page_size must be less than or equal to 100")
	}

	return nil
}
