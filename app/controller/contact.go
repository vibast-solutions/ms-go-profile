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

const contactDOBLayout = "2006-01-02"

type ContactController struct {
	contactService *service.ContactService
	logger         logrus.FieldLogger
}

func NewContactController(contactService *service.ContactService) *ContactController {
	return &ContactController{
		contactService: contactService,
		logger:         factory.NewModuleLogger("contact-controller"),
	}
}

func (c *ContactController) Create(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewCreateContactRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create contact request from context")
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request body"})
	}
	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}

	l = factory.LoggerWithContext(l, ctx).WithField("profile_id", req.GetProfileId())
	l.Info("Create contact request received")

	contact, err := c.contactService.Create(ctx.Request().Context(), req)
	if err != nil {
		l.WithError(err).Error("Create contact failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	l.WithField("contact_id", contact.ID).Info("Contact created")
	return ctx.JSON(http.StatusCreated, toContactResponse(contact))
}

func (c *ContactController) GetByID(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewGetContactRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create get contact request from context")
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request"})
	}
	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}

	l = factory.LoggerWithContext(l, ctx).WithField("contact_id", req.GetId())
	l.Info("Get contact request received")

	contact, err := c.contactService.GetByID(ctx.Request().Context(), req.GetId())
	if err != nil {
		if errors.Is(err, service.ErrContactNotFound) {
			return ctx.JSON(http.StatusNotFound, httpdto.ErrorResponse{Error: "contact not found"})
		}
		l.WithError(err).Error("Get contact failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	return ctx.JSON(http.StatusOK, toContactResponse(contact))
}

func (c *ContactController) Update(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewUpdateContactRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create update contact request from context")
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request"})
	}
	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}

	l = factory.LoggerWithContext(l, ctx).WithField("contact_id", req.GetId())
	l.Info("Update contact request received")

	contact, err := c.contactService.Update(ctx.Request().Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrContactNotFound) {
			return ctx.JSON(http.StatusNotFound, httpdto.ErrorResponse{Error: "contact not found"})
		}
		l.WithError(err).Error("Update contact failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	l.Info("Contact updated")
	return ctx.JSON(http.StatusOK, toContactResponse(contact))
}

func (c *ContactController) Delete(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewDeleteContactRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create delete contact request from context")
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request"})
	}
	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}

	l = factory.LoggerWithContext(l, ctx).WithField("contact_id", req.GetId())
	l.Info("Delete contact request received")

	if err = c.contactService.Delete(ctx.Request().Context(), req.GetId()); err != nil {
		if errors.Is(err, service.ErrContactNotFound) {
			return ctx.JSON(http.StatusNotFound, httpdto.ErrorResponse{Error: "contact not found"})
		}
		l.WithError(err).Error("Delete contact failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	l.Info("Contact deleted")
	return ctx.JSON(http.StatusOK, httpdto.DeleteResponse{Message: "contact deleted successfully"})
}

func (c *ContactController) List(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewListContactsRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create list contacts request from context")
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
	l.Info("List contacts request received")

	result, err := c.contactService.List(ctx.Request().Context(), req)
	if err != nil {
		l.WithError(err).Error("List contacts failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	contacts := make([]*types.ContactResponse, 0, len(result.Contacts))
	for _, contact := range result.Contacts {
		contacts = append(contacts, toContactResponse(contact))
	}

	return ctx.JSON(http.StatusOK, &types.ListContactsResponse{
		Contacts: contacts,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	})
}

func toContactResponse(c *entity.Contact) *types.ContactResponse {
	dob := ""
	if c.DOB != nil {
		dob = c.DOB.Format(contactDOBLayout)
	}

	return &types.ContactResponse{
		Id:        c.ID,
		FirstName: c.FirstName,
		LastName:  c.LastName,
		Nin:       c.NIN,
		Dob:       dob,
		Phone:     c.Phone,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
		UpdatedAt: c.UpdatedAt.Format(time.RFC3339),
		ProfileId: c.ProfileID,
		Type:      c.Type,
	}
}
