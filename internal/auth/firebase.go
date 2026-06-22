//go:build gcp

package auth

import (
	"context"
	"log/slog"
	"strings"

	firebase "firebase.google.com/go/v4"
	fbauth "firebase.google.com/go/v4/auth"
)

// FirebaseVerifier valida Firebase ID tokens via Admin SDK.
type FirebaseVerifier struct {
	client *fbauth.Client
}

// NovoFirebaseVerifier cria o verificador usando ADC (Application Default Credentials).
func NovoFirebaseVerifier(ctx context.Context) (*FirebaseVerifier, error) {
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, err
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}
	return &FirebaseVerifier{client: client}, nil
}

func (v *FirebaseVerifier) Verify(ctx context.Context, idToken string) *User {
	if idToken == "" {
		return nil
	}
	// Remove "Bearer " se presente
	idToken = strings.TrimPrefix(idToken, "Bearer ")
	idToken = strings.TrimPrefix(idToken, "bearer ")
	if idToken == "" {
		return nil
	}

	token, err := v.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		slog.Debug("firebase auth: token inválido", "erro", err)
		return nil
	}

	email, _ := token.Claims["email"].(string)
	return &User{
		UID:   token.UID,
		Email: email,
	}
}
