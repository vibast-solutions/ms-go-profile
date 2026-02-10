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

type AddressController struct {
	addressService *service.AddressService
	logger         logrus.FieldLogger
}

func NewAddressController(addressService *service.AddressService) *AddressController {
	return &AddressController{
		addressService: addressService,
		logger:         factory.NewModuleLogger("address-controller"),
	}
}

func (c *AddressController) Create(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewCreateAddressRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create address request from context")
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request body"})
	}
	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}

	l = factory.LoggerWithContext(l, ctx).WithField("profile_id", req.GetProfileId())
	l.Info("Create address request received")

	address, err := c.addressService.Create(ctx.Request().Context(), req)
	if err != nil {
		l.WithError(err).Error("Create address failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	l.WithField("address_id", address.ID).Info("Address created")
	return ctx.JSON(http.StatusCreated, toAddressResponse(address))
}

func (c *AddressController) GetByID(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewGetAddressRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create get address request from context")
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request"})
	}
	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}

	l = factory.LoggerWithContext(l, ctx).WithField("address_id", req.GetId())
	l.Info("Get address request received")

	address, err := c.addressService.GetByID(ctx.Request().Context(), req.GetId())
	if err != nil {
		if errors.Is(err, service.ErrAddressNotFound) {
			return ctx.JSON(http.StatusNotFound, httpdto.ErrorResponse{Error: "address not found"})
		}
		l.WithError(err).Error("Get address failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	return ctx.JSON(http.StatusOK, toAddressResponse(address))
}

func (c *AddressController) Update(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewUpdateAddressRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create update address request from context")
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request"})
	}
	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}

	l = factory.LoggerWithContext(l, ctx).WithField("address_id", req.GetId())
	l.Info("Update address request received")

	address, err := c.addressService.Update(ctx.Request().Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrAddressNotFound) {
			return ctx.JSON(http.StatusNotFound, httpdto.ErrorResponse{Error: "address not found"})
		}
		l.WithError(err).Error("Update address failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	l.Info("Address updated")
	return ctx.JSON(http.StatusOK, toAddressResponse(address))
}

func (c *AddressController) Delete(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewDeleteAddressRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create delete address request from context")
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request"})
	}
	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}

	l = factory.LoggerWithContext(l, ctx).WithField("address_id", req.GetId())
	l.Info("Delete address request received")

	if err = c.addressService.Delete(ctx.Request().Context(), req.GetId()); err != nil {
		if errors.Is(err, service.ErrAddressNotFound) {
			return ctx.JSON(http.StatusNotFound, httpdto.ErrorResponse{Error: "address not found"})
		}
		l.WithError(err).Error("Delete address failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	l.Info("Address deleted")
	return ctx.JSON(http.StatusOK, httpdto.DeleteResponse{Message: "address deleted successfully"})
}

func (c *AddressController) List(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewListAddressesRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create list addresses request from context")
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request"})
	}
	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}

	l = factory.LoggerWithContext(l, ctx).WithFields(logrus.Fields{
		"profile_id": req.GetProfileId(),
		"page":       req.GetPage(),
		"page_size":  req.GetPageSize(),
	})
	l.Info("List addresses request received")

	result, err := c.addressService.List(ctx.Request().Context(), req)
	if err != nil {
		l.WithError(err).Error("List addresses failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	addresses := make([]*types.AddressResponse, 0, len(result.Addresses))
	for _, address := range result.Addresses {
		addresses = append(addresses, toAddressResponse(address))
	}

	return ctx.JSON(http.StatusOK, &types.ListAddressesResponse{
		Addresses: addresses,
		Page:      result.Page,
		PageSize:  result.PageSize,
		Total:     result.Total,
	})
}

func toAddressResponse(a *entity.Address) *types.AddressResponse {
	return &types.AddressResponse{
		Id:             a.ID,
		StreetName:     a.StreetName,
		StreenNo:       a.StreenNo,
		City:           a.City,
		County:         a.County,
		Country:        a.Country,
		ProfileId:      a.ProfileID,
		PostalCode:     a.PostalCode,
		Building:       a.Building,
		Apartment:      a.Apartment,
		AdditionalData: a.AdditionalData,
		Type:           a.Type,
		CreatedAt:      a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      a.UpdatedAt.Format(time.RFC3339),
	}
}
