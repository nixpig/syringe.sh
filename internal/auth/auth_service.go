package auth

import (
	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/pkg/validation"
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
	validate validation.Validator
	logger   *zerolog.Logger
}

func NewAuthService(
	store AuthStore,
	validate validation.Validator,
) AuthServiceImpl {
	return AuthServiceImpl{
		store:    store,
		validate: validate,
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
