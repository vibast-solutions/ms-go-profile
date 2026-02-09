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

type ProfileController struct {
	profileService *service.ProfileService

	logger logrus.FieldLogger
}

func NewProfileController(profileService *service.ProfileService) *ProfileController {
	return &ProfileController{
		profileService: profileService,
		logger:         factory.NewModuleLogger("profile-controller"),
	}
}

func (c *ProfileController) Create(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewCreateProfileRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create profile request from context")

		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request body"})
	}

	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}
	l = factory.LoggerWithContext(l, ctx).WithField("user_id", req.GetUserId())
	l.Info("Create profile request received")

	profile, err := c.profileService.Create(ctx.Request().Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrProfileAlreadyExists) {
			return ctx.JSON(http.StatusConflict, httpdto.ErrorResponse{Error: "profile already exists for this user"})
		}

		l.WithError(err).Error("Create profile failed")

		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	l.Info("Profile created")

	return ctx.JSON(http.StatusCreated, toProfileResponse(profile))
}

func (c *ProfileController) GetByID(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewGetProfileRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Failed to create profile request from context")

		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request"})
	}

	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}

	l = factory.LoggerWithContext(l, ctx).WithField("profile_id", req.GetId())
	l.Info("Get profile request received")

	profile, err := c.profileService.GetByID(ctx.Request().Context(), req.GetId())
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			return ctx.JSON(http.StatusNotFound, httpdto.ErrorResponse{Error: "profile not found"})
		}
		l.WithError(err).Error("Get profile failed")

		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	return ctx.JSON(http.StatusOK, toProfileResponse(profile))
}

func (c *ProfileController) GetByUserID(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewGetProfileByUserIDRequestFromContext(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request"})
	}

	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}
	l = factory.LoggerWithContext(l, ctx).WithField("user_id", req.GetUserId())
	l.Info("Get profile by user ID request received")

	profile, err := c.profileService.GetByUserID(ctx.Request().Context(), req.GetUserId())
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			l.WithField("user_id", req.GetUserId()).Warn("Get profile by user ID failed: not found")

			return ctx.JSON(http.StatusNotFound, httpdto.ErrorResponse{Error: "profile not found"})
		}

		l.WithError(err).WithField("user_id", req.GetUserId()).Error("Get profile by user ID failed")

		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	return ctx.JSON(http.StatusOK, toProfileResponse(profile))
}

func (c *ProfileController) Update(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewUpdateProfileRequestFromContext(ctx)
	if err != nil {
		l.WithError(err).Debug("Invalid profile ID")

		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request"})
	}

	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}
	l = factory.LoggerWithContext(l, ctx).WithField("profile_id", req.GetId())
	l.Info("Update profile request received")

	profile, err := c.profileService.Update(ctx.Request().Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			return ctx.JSON(http.StatusNotFound, httpdto.ErrorResponse{Error: "profile not found"})
		}
		l.WithError(err).Error("Update profile failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	l.Info("Profile updated")

	return ctx.JSON(http.StatusOK, toProfileResponse(profile))
}

func (c *ProfileController) Delete(ctx echo.Context) error {
	l := c.logger
	req, err := types.NewDeleteProfileRequestFromContext(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: "invalid request"})
	}

	if err = req.Validate(); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpdto.ErrorResponse{Error: err.Error()})
	}
	l = factory.LoggerWithContext(l, ctx).WithField("profile_id", req.GetId())

	l.Info("Delete profile request received")
	if err := c.profileService.Delete(ctx.Request().Context(), req.GetId()); err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			return ctx.JSON(http.StatusNotFound, httpdto.ErrorResponse{Error: "profile not found"})
		}
		l.WithError(err).Error("Delete profile failed")
		return ctx.JSON(http.StatusInternalServerError, httpdto.ErrorResponse{Error: "internal server error"})
	}

	l.Info("Profile deleted")
	return ctx.JSON(http.StatusOK, httpdto.DeleteResponse{Message: "profile deleted successfully"})
}

func toProfileResponse(p *entity.Profile) *types.ProfileResponse {
	return &types.ProfileResponse{
		Id:        p.ID,
		UserId:    p.UserID,
		Email:     p.Email,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}
