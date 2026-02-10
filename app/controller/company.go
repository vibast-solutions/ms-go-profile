package controller

import (
	"errors"
	"net/http"
	"time"

	httpdto "github.com/vibast-solutions/ms-go-profile/app/dto"
	"github.com/vibast-solutions/ms-go-profile/app/entity"
	"github.com/vibast-solutions/ms-go-profile/app/factory"
	"github.com/vibast-solutions/ms-go-profile/app/service"
	"github.com/vibast-solutions/ms-go-profile/app/types"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type CompanyController struct {
	companyService *service.CompanyService
	logger         logrus.FieldLogger
}

func NewCompanyController(companyService *service.CompanyService) *CompanyController {
	return &CompanyController{
		companyService: companyService,
		logger:         factory.NewModuleLogger("company-controller"),
	}
}

func (c *CompanyController) Create(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewCreateCompanyRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create company request from context")
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request body"})
	}
	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}

	l = factory.LoggerWithContext(l, ctx).WithField("profile_id", req.GetProfileId())
	l.Info("Create company request received")

	company, err := c.companyService.Create(ctx.Request().Context(), req)
	if err != nil {
		l.WithError(err).Error("Create company failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	l.WithField("company_id", company.ID).Info("Company created")
	return ctx.JSON(http.StatusCreated, toCompanyResponse(company))
}

func (c *CompanyController) GetByID(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewGetCompanyRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create get company request from context")
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request"})
	}
	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}

	l = factory.LoggerWithContext(l, ctx).WithField("company_id", req.GetId())
	l.Info("Get company request received")

	company, err := c.companyService.GetByID(ctx.Request().Context(), req.GetId())
	if err != nil {
		if errors.Is(err, service.ErrCompanyNotFound) {
			return ctx.JSON(http.StatusNotFound, httpdto.ErrorResponse{Error: "company not found"})
		}
		l.WithError(err).Error("Get company failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	return ctx.JSON(http.StatusOK, toCompanyResponse(company))
}

func (c *CompanyController) Update(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewUpdateCompanyRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create update company request from context")
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request"})
	}
	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}

	l = factory.LoggerWithContext(l, ctx).WithField("company_id", req.GetId())
	l.Info("Update company request received")

	company, err := c.companyService.Update(ctx.Request().Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrCompanyNotFound) {
			return ctx.JSON(http.StatusNotFound, httpdto.ErrorResponse{Error: "company not found"})
		}
		l.WithError(err).Error("Update company failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	l.Info("Company updated")
	return ctx.JSON(http.StatusOK, toCompanyResponse(company))
}

func (c *CompanyController) Delete(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewDeleteCompanyRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create delete company request from context")
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request"})
	}
	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}

	l = factory.LoggerWithContext(l, ctx).WithField("company_id", req.GetId())
	l.Info("Delete company request received")

	if err = c.companyService.Delete(ctx.Request().Context(), req.GetId()); err != nil {
		if errors.Is(err, service.ErrCompanyNotFound) {
			return ctx.JSON(http.StatusNotFound, httpdto.ErrorResponse{Error: "company not found"})
		}
		l.WithError(err).Error("Delete company failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	l.Info("Company deleted")
	return ctx.JSON(http.StatusOK, httpdto.DeleteResponse{Message: "company deleted successfully"})
}

func (c *CompanyController) List(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewListCompaniesRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create list companies request from context")
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request"})
	}
	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}

	l = factory.LoggerWithContext(l, ctx).WithFields(logrus.Fields{
		"profile_id": req.GetProfileId(),
		"page":       req.GetPage(),
		"page_size":  req.GetPageSize(),
		"type":       req.GetType(),
	})
	l.Info("List companies request received")

	result, err := c.companyService.List(ctx.Request().Context(), req)
	if err != nil {
		l.WithError(err).Error("List companies failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	companies := make([]*types.CompanyResponse, 0, len(result.Companies))
	for _, company := range result.Companies {
		companies = append(companies, toCompanyResponse(company))
	}

	return ctx.JSON(http.StatusOK, &types.ListCompaniesResponse{
		Companies: companies,
		Page:      result.Page,
		PageSize:  result.PageSize,
		Total:     result.Total,
	})
}

func toCompanyResponse(company *entity.Company) *types.CompanyResponse {
	return &types.CompanyResponse{
		Id:             company.ID,
		Name:           company.Name,
		RegistrationNo: company.RegistrationNo,
		FiscalCode:     company.FiscalCode,
		ProfileId:      company.ProfileID,
		Type:           company.Type,
		CreatedAt:      company.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      company.UpdatedAt.Format(time.RFC3339),
	}
}
