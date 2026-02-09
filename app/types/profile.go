package types

import (
	"errors"
	"strconv"

	"github.com/labstack/echo/v4"
)

func NewCreateProfileRequestFromContext(ctx echo.Context) (*CreateProfileRequest, error) {
	var body CreateProfileRequest
	err := ctx.Bind(&body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

func (r *CreateProfileRequest) Validate() error {
	if r.UserId == 0 {
		return errors.New("user_id is required")
	}
	if r.Email == "" {
		return errors.New("email is required")
	}

	return nil
}

func NewGetProfileRequestFromContext(ctx echo.Context) (*GetProfileRequest, error) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	return &GetProfileRequest{Id: id}, nil
}

func (r *GetProfileRequest) Validate() error {
	if r.Id == 0 {
		return errors.New("invalid id provided")
	}

	return nil
}

func NewGetProfileByUserIDRequestFromContext(ctx echo.Context) (*GetProfileByUserIDRequest, error) {
	userID, err := strconv.ParseUint(ctx.Param("user_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	return &GetProfileByUserIDRequest{UserId: userID}, nil
}

func (r *GetProfileByUserIDRequest) Validate() error {
	if r.UserId == 0 {
		return errors.New("invalid user id provided")
	}

	return nil
}

func NewUpdateProfileRequestFromContext(ctx echo.Context) (*UpdateProfileRequest, error) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	var body UpdateProfileRequest
	err = ctx.Bind(&body)
	if err != nil {
		return nil, err
	}

	body.Id = id

	return &body, nil
}

func (r *UpdateProfileRequest) Validate() error {
	if r.Id == 0 {
		return errors.New("invalid id provided")
	}

	if r.Email == "" {
		return errors.New("invalid email")
	}

	//TODO: properly validate email

	return nil
}

func NewDeleteProfileRequestFromContext(ctx echo.Context) (*DeleteProfileRequest, error) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	return &DeleteProfileRequest{Id: id}, nil
}

func (r *DeleteProfileRequest) Validate() error {
	if r.Id == 0 {
		return errors.New("invalid id provided")
	}

	return nil
}
