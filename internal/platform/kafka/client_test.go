package kafka

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go/sasl"
	"golang.org/x/oauth2"
)

func TestClientConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  ClientConfig
		wantErr bool
	}{
		{
			name: "local plaintext works by default",
			config: ClientConfig{
				Brokers: []string{"localhost:9092"},
			},
		},
		{
			name: "plain auth requires tls",
			config: ClientConfig{
				Brokers:      []string{"broker:9092"},
				AuthMode:     AuthModePlain,
				SASLUsername: "service-account@example.iam.gserviceaccount.com",
				SASLPassword: "token",
			},
			wantErr: true,
		},
		{
			name: "plain auth requires username and password",
			config: ClientConfig{
				Brokers:    []string{"broker:9092"},
				AuthMode:   AuthModePlain,
				TLSEnabled: true,
			},
			wantErr: true,
		},
		{
			name: "google access token auth requires principal email",
			config: ClientConfig{
				Brokers:    []string{"broker:9092"},
				AuthMode:   AuthModeGoogleAccessToken,
				TLSEnabled: true,
			},
			wantErr: true,
		},
		{
			name: "google access token auth validates",
			config: ClientConfig{
				Brokers:           []string{"broker:9092"},
				AuthMode:          AuthModeGoogleAccessToken,
				TLSEnabled:        true,
				GCPPrincipalEmail: "ingestion-api-dev@example.iam.gserviceaccount.com",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.config.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestGoogleAccessTokenMechanismUsesFreshToken(t *testing.T) {
	t.Parallel()

	mechanism := newGoogleAccessTokenMechanism("svc@example.iam.gserviceaccount.com", &stubTokenSource{
		token: &oauth2.Token{AccessToken: "test-access-token"},
	})

	sess, ir, err := mechanism.Start(context.Background())
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if string(ir) != "\x00svc@example.iam.gserviceaccount.com\x00test-access-token" {
		t.Fatalf("unexpected initial response %q", string(ir))
	}
	if sess == nil {
		t.Fatalf("expected a non-nil state machine")
	}
}

func TestGoogleAccessTokenMechanismPropagatesTokenErrors(t *testing.T) {
	t.Parallel()

	var mechanism sasl.Mechanism = newGoogleAccessTokenMechanism("svc@example.iam.gserviceaccount.com", &stubTokenSource{
		err: errors.New("boom"),
	})

	if _, _, err := mechanism.Start(context.Background()); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

type stubTokenSource struct {
	token *oauth2.Token
	err   error
}

func (s *stubTokenSource) Token() (*oauth2.Token, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.token, nil
}
