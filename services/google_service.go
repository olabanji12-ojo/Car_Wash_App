package services

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
)

type GoogleService struct {
	Verifier *oidc.IDTokenVerifier
	Provider *oidc.Provider
}

func NewGoogleService(ctx context.Context, clientID string) (*GoogleService, error) {
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("oidc provider: %w", err)
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &GoogleService{
		Verifier: verifier,
		Provider: provider,
	}, nil
}

// VerifyIDToken verifies raw ID token and returns structured claims.
type GoogleClaims struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (s *GoogleService) VerifyIDToken(ctx context.Context, rawToken string) (*GoogleClaims, error) {
	idToken, err := s.Verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, err
	}
	var claims GoogleClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}
	return &claims, nil
}
