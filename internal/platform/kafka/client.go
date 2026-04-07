package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	kafkaGo "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	DefaultDialTimeout            = 10 * time.Second
	DefaultGoogleAccessTokenScope = "https://www.googleapis.com/auth/cloud-platform"
)

type AuthMode string

const (
	AuthModeNone              AuthMode = "none"
	AuthModePlain             AuthMode = "plain"
	AuthModeGoogleAccessToken AuthMode = "google_access_token"
)

type ClientConfig struct {
	Brokers               []string
	ClientID              string
	DialTimeout           time.Duration
	TLSEnabled            bool
	TLSInsecureSkipVerify bool
	TLSServerName         string
	TLSCAFile             string
	AuthMode              AuthMode
	SASLUsername          string
	SASLPassword          string
	GCPPrincipalEmail     string
	GCPAccessTokenScope   string
}

func (c ClientConfig) Validate() error {
	if len(c.Brokers) == 0 {
		return errors.New("at least one Kafka broker is required")
	}

	switch c.normalizedAuthMode() {
	case AuthModeNone:
		return nil
	case AuthModePlain:
		if !c.TLSEnabled {
			return errors.New("KAFKA_TLS_ENABLED must be true when KAFKA_AUTH_MODE=plain")
		}
		if strings.TrimSpace(c.SASLUsername) == "" {
			return errors.New("KAFKA_SASL_USERNAME is required when KAFKA_AUTH_MODE=plain")
		}
		if strings.TrimSpace(c.SASLPassword) == "" {
			return errors.New("KAFKA_SASL_PASSWORD is required when KAFKA_AUTH_MODE=plain")
		}
		return nil
	case AuthModeGoogleAccessToken:
		if !c.TLSEnabled {
			return errors.New("KAFKA_TLS_ENABLED must be true when KAFKA_AUTH_MODE=google_access_token")
		}
		if strings.TrimSpace(c.GCPPrincipalEmail) == "" {
			return errors.New("KAFKA_GCP_PRINCIPAL_EMAIL is required when KAFKA_AUTH_MODE=google_access_token")
		}
		return nil
	default:
		return fmt.Errorf("unsupported KAFKA_AUTH_MODE %q", c.AuthMode)
	}
}

func NewTransport(cfg ClientConfig) (*kafkaGo.Transport, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	tlsConfig, err := cfg.tlsConfig()
	if err != nil {
		return nil, err
	}

	mechanism, err := cfg.saslMechanism()
	if err != nil {
		return nil, err
	}

	return &kafkaGo.Transport{
		ClientID:    cfg.ClientID,
		DialTimeout: cfg.effectiveDialTimeout(),
		TLS:         tlsConfig,
		SASL:        mechanism,
	}, nil
}

func NewDialer(cfg ClientConfig) (*kafkaGo.Dialer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	tlsConfig, err := cfg.tlsConfig()
	if err != nil {
		return nil, err
	}

	mechanism, err := cfg.saslMechanism()
	if err != nil {
		return nil, err
	}

	return &kafkaGo.Dialer{
		ClientID:      cfg.ClientID,
		Timeout:       cfg.effectiveDialTimeout(),
		DualStack:     true,
		TLS:           tlsConfig,
		SASLMechanism: mechanism,
	}, nil
}

func (c ClientConfig) effectiveDialTimeout() time.Duration {
	if c.DialTimeout > 0 {
		return c.DialTimeout
	}
	return DefaultDialTimeout
}

func (c ClientConfig) normalizedAuthMode() AuthMode {
	mode := AuthMode(strings.TrimSpace(string(c.AuthMode)))
	if mode == "" {
		return AuthModeNone
	}
	return mode
}

func (c ClientConfig) tlsConfig() (*tls.Config, error) {
	if !c.TLSEnabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: c.TLSInsecureSkipVerify,
		ServerName:         strings.TrimSpace(c.TLSServerName),
	}

	if strings.TrimSpace(c.TLSCAFile) == "" {
		return tlsConfig, nil
	}

	caBundle, err := os.ReadFile(c.TLSCAFile)
	if err != nil {
		return nil, fmt.Errorf("read Kafka CA bundle %q: %w", c.TLSCAFile, err)
	}

	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("load system certificate pool: %w", err)
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if !rootCAs.AppendCertsFromPEM(caBundle) {
		return nil, fmt.Errorf("parse Kafka CA bundle %q", c.TLSCAFile)
	}

	tlsConfig.RootCAs = rootCAs
	return tlsConfig, nil
}

func (c ClientConfig) saslMechanism() (sasl.Mechanism, error) {
	switch c.normalizedAuthMode() {
	case AuthModeNone:
		return nil, nil
	case AuthModePlain:
		return plain.Mechanism{
			Username: c.SASLUsername,
			Password: c.SASLPassword,
		}, nil
	case AuthModeGoogleAccessToken:
		scope := strings.TrimSpace(c.GCPAccessTokenScope)
		if scope == "" {
			scope = DefaultGoogleAccessTokenScope
		}

		tokenSource, err := google.DefaultTokenSource(context.Background(), scope)
		if err != nil {
			return nil, fmt.Errorf("build Google ADC token source: %w", err)
		}

		return newGoogleAccessTokenMechanism(c.GCPPrincipalEmail, tokenSource), nil
	default:
		return nil, fmt.Errorf("unsupported KAFKA_AUTH_MODE %q", c.AuthMode)
	}
}

type oauthTokenSource interface {
	Token() (*oauth2.Token, error)
}

type googleAccessTokenMechanism struct {
	username    string
	tokenSource oauthTokenSource
}

func newGoogleAccessTokenMechanism(username string, tokenSource oauthTokenSource) sasl.Mechanism {
	return googleAccessTokenMechanism{
		username:    strings.TrimSpace(username),
		tokenSource: tokenSource,
	}
}

func (m googleAccessTokenMechanism) Name() string {
	return "PLAIN"
}

func (m googleAccessTokenMechanism) Start(ctx context.Context) (sasl.StateMachine, []byte, error) {
	token, err := m.tokenSource.Token()
	if err != nil {
		return nil, nil, fmt.Errorf("obtain Google access token: %w", err)
	}
	if token == nil || strings.TrimSpace(token.AccessToken) == "" {
		return nil, nil, errors.New("empty Google access token")
	}

	initialResponse := []byte(fmt.Sprintf("\x00%s\x00%s", m.username, token.AccessToken))
	return plainStateMachine{}, initialResponse, nil
}

type plainStateMachine struct{}

func (plainStateMachine) Next(ctx context.Context, challenge []byte) (bool, []byte, error) {
	return true, nil, nil
}
