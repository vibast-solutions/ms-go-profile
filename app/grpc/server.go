package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/vibast-solutions/ms-go-profile/app/entity"
	"github.com/vibast-solutions/ms-go-profile/app/service"
	"github.com/vibast-solutions/ms-go-profile/app/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProfileServer struct {
	types.UnimplementedProfileServiceServer
	profileService *service.ProfileService
	contactService *service.ContactService
	addressService *service.AddressService
}

const grpcContactDOBLayout = "2006-01-02"

func NewProfileServer(profileService *service.ProfileService, contactService *service.ContactService, addressService *service.AddressService) *ProfileServer {
	return &ProfileServer{
		profileService: profileService,
		contactService: contactService,
		addressService: addressService,
	}
}

func (s *ProfileServer) CreateProfile(ctx context.Context, pbReq *types.CreateProfileRequest) (*types.ProfileResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.WithField("user_id", pbReq.UserId).Debug("Create profile validation failed (grpc)")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("user_id", pbReq.GetUserId()).Info("Create profile request received (grpc)")
	profile, err := s.profileService.Create(ctx, pbReq)
	if err != nil {
		if errors.Is(err, service.ErrProfileAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "profile already exists for this user")
		}
		l.WithError(err).WithField("user_id", pbReq.GetUserId()).Error("Create profile failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	l.WithField("profile_id", profile.ID).WithField("user_id", profile.UserID).Info("Profile created (grpc)")

	return &types.ProfileResponse{
		Id:        profile.ID,
		UserId:    profile.UserID,
		Email:     profile.Email,
		CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt: profile.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ProfileServer) GetProfile(ctx context.Context, pbReq *types.GetProfileRequest) (*types.ProfileResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("Get profile validation failed (grpc)")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("profile_id", pbReq.GetId()).Info("Get profile request received (grpc)")
	profile, err := s.profileService.GetByID(ctx, pbReq.GetId())
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		l.WithError(err).WithField("profile_id", pbReq.GetId()).Error("Get profile failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &types.ProfileResponse{
		Id:        profile.ID,
		UserId:    profile.UserID,
		Email:     profile.Email,
		CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt: profile.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ProfileServer) GetProfileByUserID(ctx context.Context, pbReq *types.GetProfileByUserIDRequest) (*types.ProfileResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("Get profile by user ID validation failed (grpc)")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("user_id", pbReq.GetUserId()).Info("Get profile by user ID request received (grpc)")
	profile, err := s.profileService.GetByUserID(ctx, pbReq.GetUserId())
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		l.WithError(err).WithField("user_id", pbReq.GetUserId()).Error("Get profile by user ID failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &types.ProfileResponse{
		Id:        profile.ID,
		UserId:    profile.UserID,
		Email:     profile.Email,
		CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt: profile.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ProfileServer) UpdateProfile(ctx context.Context, pbReq *types.UpdateProfileRequest) (*types.ProfileResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("Update profile validation failed (grpc)")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("profile_id", pbReq.GetId()).Info("Update profile request received (grpc)")
	profile, err := s.profileService.Update(ctx, pbReq)
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		l.WithError(err).WithField("profile_id", pbReq.GetId()).Error("Update profile failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	l.WithField("profile_id", pbReq.GetId()).Info("Profile updated (grpc)")
	return &types.ProfileResponse{
		Id:        profile.ID,
		UserId:    profile.UserID,
		Email:     profile.Email,
		CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt: profile.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ProfileServer) DeleteProfile(ctx context.Context, pbReq *types.DeleteProfileRequest) (*types.DeleteProfileResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("Delete profile validation failed (grpc)")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("profile_id", pbReq.GetId()).Info("Delete profile request received (grpc)")
	if err := s.profileService.Delete(ctx, pbReq.GetId()); err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		l.WithError(err).WithField("profile_id", pbReq.GetId()).Error("Delete profile failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	l.WithField("profile_id", pbReq.GetId()).Info("Profile deleted (grpc)")
	return &types.DeleteProfileResponse{
		Message: "profile deleted successfully",
	}, nil
}

func (s *ProfileServer) CreateContact(ctx context.Context, pbReq *types.CreateContactRequest) (*types.ContactResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.WithField("profile_id", pbReq.GetProfileId()).Debug("Create contact validation failed (grpc)")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("profile_id", pbReq.GetProfileId()).Info("Create contact request received (grpc)")
	contact, err := s.contactService.Create(ctx, pbReq)
	if err != nil {
		l.WithError(err).WithField("profile_id", pbReq.GetProfileId()).Error("Create contact failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	l.WithField("contact_id", contact.ID).Info("Contact created (grpc)")
	return &types.ContactResponse{
		Id:        contact.ID,
		FirstName: contact.FirstName,
		LastName:  contact.LastName,
		Nin:       contact.NIN,
		Dob:       contactDOBString(contact.DOB),
		Phone:     contact.Phone,
		CreatedAt: contact.CreatedAt.Format(time.RFC3339),
		UpdatedAt: contact.UpdatedAt.Format(time.RFC3339),
		ProfileId: contact.ProfileID,
		Type:      contact.Type,
	}, nil
}

func (s *ProfileServer) GetContact(ctx context.Context, pbReq *types.GetContactRequest) (*types.ContactResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("Get contact validation failed (grpc)")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("contact_id", pbReq.GetId()).Info("Get contact request received (grpc)")
	contact, err := s.contactService.GetByID(ctx, pbReq.GetId())
	if err != nil {
		if errors.Is(err, service.ErrContactNotFound) {
			return nil, status.Error(codes.NotFound, "contact not found")
		}
		l.WithError(err).WithField("contact_id", pbReq.GetId()).Error("Get contact failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &types.ContactResponse{
		Id:        contact.ID,
		FirstName: contact.FirstName,
		LastName:  contact.LastName,
		Nin:       contact.NIN,
		Dob:       contactDOBString(contact.DOB),
		Phone:     contact.Phone,
		CreatedAt: contact.CreatedAt.Format(time.RFC3339),
		UpdatedAt: contact.UpdatedAt.Format(time.RFC3339),
		ProfileId: contact.ProfileID,
		Type:      contact.Type,
	}, nil
}

func (s *ProfileServer) UpdateContact(ctx context.Context, pbReq *types.UpdateContactRequest) (*types.ContactResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("Update contact validation failed (grpc)")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("contact_id", pbReq.GetId()).Info("Update contact request received (grpc)")
	contact, err := s.contactService.Update(ctx, pbReq)
	if err != nil {
		if errors.Is(err, service.ErrContactNotFound) {
			return nil, status.Error(codes.NotFound, "contact not found")
		}
		l.WithError(err).WithField("contact_id", pbReq.GetId()).Error("Update contact failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	l.WithField("contact_id", pbReq.GetId()).Info("Contact updated (grpc)")
	return &types.ContactResponse{
		Id:        contact.ID,
		FirstName: contact.FirstName,
		LastName:  contact.LastName,
		Nin:       contact.NIN,
		Dob:       contactDOBString(contact.DOB),
		Phone:     contact.Phone,
		CreatedAt: contact.CreatedAt.Format(time.RFC3339),
		UpdatedAt: contact.UpdatedAt.Format(time.RFC3339),
		ProfileId: contact.ProfileID,
		Type:      contact.Type,
	}, nil
}

func (s *ProfileServer) DeleteContact(ctx context.Context, pbReq *types.DeleteContactRequest) (*types.DeleteContactResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("Delete contact validation failed (grpc)")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("contact_id", pbReq.GetId()).Info("Delete contact request received (grpc)")
	if err := s.contactService.Delete(ctx, pbReq.GetId()); err != nil {
		if errors.Is(err, service.ErrContactNotFound) {
			return nil, status.Error(codes.NotFound, "contact not found")
		}
		l.WithError(err).WithField("contact_id", pbReq.GetId()).Error("Delete contact failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	l.WithField("contact_id", pbReq.GetId()).Info("Contact deleted (grpc)")
	return &types.DeleteContactResponse{
		Message: "contact deleted successfully",
	}, nil
}

func (s *ProfileServer) ListContacts(ctx context.Context, pbReq *types.ListContactsRequest) (*types.ListContactsResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("List contacts validation failed (grpc)")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithFields(map[string]interface{}{
		"profile_id": pbReq.GetProfileId(),
		"page":       pbReq.GetPage(),
		"page_size":  pbReq.GetPageSize(),
	}).Info("List contacts request received (grpc)")

	result, err := s.contactService.List(ctx, pbReq)
	if err != nil {
		l.WithError(err).Error("List contacts failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	contacts := make([]*types.ContactResponse, 0, len(result.Contacts))
	for _, contact := range result.Contacts {
		contacts = append(contacts, &types.ContactResponse{
			Id:        contact.ID,
			FirstName: contact.FirstName,
			LastName:  contact.LastName,
			Nin:       contact.NIN,
			Dob:       contactDOBString(contact.DOB),
			Phone:     contact.Phone,
			CreatedAt: contact.CreatedAt.Format(time.RFC3339),
			UpdatedAt: contact.UpdatedAt.Format(time.RFC3339),
			ProfileId: contact.ProfileID,
			Type:      contact.Type,
		})
	}

	return &types.ListContactsResponse{
		Contacts: contacts,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	}, nil
}

func (s *ProfileServer) CreateAddress(ctx context.Context, pbReq *types.CreateAddressRequest) (*types.AddressResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.WithField("profile_id", pbReq.GetProfileId()).Debug("Create address validation failed (grpc)")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("profile_id", pbReq.GetProfileId()).Info("Create address request received (grpc)")
	address, err := s.addressService.Create(ctx, pbReq)
	if err != nil {
		l.WithError(err).WithField("profile_id", pbReq.GetProfileId()).Error("Create address failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	l.WithField("address_id", address.ID).Info("Address created (grpc)")
	return toAddressResponse(address), nil
}

func (s *ProfileServer) GetAddress(ctx context.Context, pbReq *types.GetAddressRequest) (*types.AddressResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("Get address validation failed (grpc)")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("address_id", pbReq.GetId()).Info("Get address request received (grpc)")
	address, err := s.addressService.GetByID(ctx, pbReq.GetId())
	if err != nil {
		if errors.Is(err, service.ErrAddressNotFound) {
			return nil, status.Error(codes.NotFound, "address not found")
		}
		l.WithError(err).WithField("address_id", pbReq.GetId()).Error("Get address failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return toAddressResponse(address), nil
}

func (s *ProfileServer) UpdateAddress(ctx context.Context, pbReq *types.UpdateAddressRequest) (*types.AddressResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("Update address validation failed (grpc)")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("address_id", pbReq.GetId()).Info("Update address request received (grpc)")
	address, err := s.addressService.Update(ctx, pbReq)
	if err != nil {
		if errors.Is(err, service.ErrAddressNotFound) {
			return nil, status.Error(codes.NotFound, "address not found")
		}
		l.WithError(err).WithField("address_id", pbReq.GetId()).Error("Update address failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	l.WithField("address_id", pbReq.GetId()).Info("Address updated (grpc)")
	return toAddressResponse(address), nil
}

func (s *ProfileServer) DeleteAddress(ctx context.Context, pbReq *types.DeleteAddressRequest) (*types.DeleteAddressResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("Delete address validation failed (grpc)")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("address_id", pbReq.GetId()).Info("Delete address request received (grpc)")
	if err := s.addressService.Delete(ctx, pbReq.GetId()); err != nil {
		if errors.Is(err, service.ErrAddressNotFound) {
			return nil, status.Error(codes.NotFound, "address not found")
		}
		l.WithError(err).WithField("address_id", pbReq.GetId()).Error("Delete address failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	l.WithField("address_id", pbReq.GetId()).Info("Address deleted (grpc)")
	return &types.DeleteAddressResponse{Message: "address deleted successfully"}, nil
}

func (s *ProfileServer) ListAddresses(ctx context.Context, pbReq *types.ListAddressesRequest) (*types.ListAddressesResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("List addresses validation failed (grpc)")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithFields(map[string]interface{}{
		"profile_id": pbReq.GetProfileId(),
		"page":       pbReq.GetPage(),
		"page_size":  pbReq.GetPageSize(),
	}).Info("List addresses request received (grpc)")

	result, err := s.addressService.List(ctx, pbReq)
	if err != nil {
		l.WithError(err).Error("List addresses failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	addresses := make([]*types.AddressResponse, 0, len(result.Addresses))
	for _, address := range result.Addresses {
		addresses = append(addresses, toAddressResponse(address))
	}

	return &types.ListAddressesResponse{
		Addresses: addresses,
		Page:      result.Page,
		PageSize:  result.PageSize,
		Total:     result.Total,
	}, nil
}

func toAddressResponse(address *entity.Address) *types.AddressResponse {
	return &types.AddressResponse{
		Id:             address.ID,
		StreetName:     address.StreetName,
		StreenNo:       address.StreenNo,
		City:           address.City,
		County:         address.County,
		Country:        address.Country,
		ProfileId:      address.ProfileID,
		PostalCode:     address.PostalCode,
		Building:       address.Building,
		Apartment:      address.Apartment,
		AdditionalData: address.AdditionalData,
		Type:           address.Type,
		CreatedAt:      address.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      address.UpdatedAt.Format(time.RFC3339),
	}
}

func contactDOBString(dob *time.Time) string {
	if dob == nil {
		return ""
	}
	return dob.Format(grpcContactDOBLayout)
}
