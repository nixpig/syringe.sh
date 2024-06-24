package auth

import (
	"github.com/charmbracelet/ssh"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

type AuthenticateUserRequest struct {
	Username  string
	PublicKey ssh.PublicKey
}

type AuthenticateUserResponse struct {
	Auth bool
}

type AuthService interface {
	AuthenticateUser(authDetails AuthenticateUserRequest) (*AuthenticateUserResponse, error)
}

type AuthServiceImpl struct {
	store    AuthStore
	validate *validator.Validate
	logger   *zerolog.Logger
}

func NewAuthService(
	store AuthStore,
	validate *validator.Validate,
	logger *zerolog.Logger,
) AuthServiceImpl {
	return AuthServiceImpl{
		store:    store,
		validate: validate,
		logger:   logger,
	}
}

func (a AuthServiceImpl) AuthenticateUser(
	authDetails AuthenticateUserRequest,
) (*AuthenticateUserResponse, error) {
	if err := a.validate.Struct(authDetails); err != nil {
		return nil, err
	}

	publicKeysDetails, err := a.store.GetUserPublicKeys(authDetails.Username)
	if err != nil {
		return nil, err
	}

	for _, v := range *publicKeysDetails {
		parsed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(v.PublicKey))
		if err != nil {
			return nil, err
		}

		if ssh.KeysEqual(
			authDetails.PublicKey,
			parsed,
		) {
			return &AuthenticateUserResponse{Auth: true}, nil
		}
	}

	return &AuthenticateUserResponse{Auth: false}, nil
}
