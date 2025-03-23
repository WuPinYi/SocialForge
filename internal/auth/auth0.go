package auth

import (
	"context"
	"fmt"
	"net/url"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Auth0Config holds the configuration for Auth0
type Auth0Config struct {
	Domain string
}

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	*validator.RegisteredClaims
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Validate implements validator.CustomClaims.
func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

// Auth0Middleware handles Auth0 authentication
type Auth0Middleware struct {
	validator *validator.Validator
}

// NewAuth0Middleware creates a new Auth0 middleware
func NewAuth0Middleware(config Auth0Config) (*Auth0Middleware, error) {
	issuerURL := fmt.Sprintf("https://%s/", config.Domain)
	issuerURLParsed, err := url.Parse(issuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse issuer URL: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURLParsed, 5)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL,
		[]string{config.Domain},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %v", err)
	}

	return &Auth0Middleware{
		validator: jwtValidator,
	}, nil
}

// UnaryInterceptor implements the gRPC unary interceptor for Auth0 authentication
func (m *Auth0Middleware) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Skip authentication for reflection service
	if info.FullMethod == "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo" {
		return handler(ctx, req)
	}

	// Get the authorization header from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	// Extract the token from the authorization header
	token := authHeader[0]
	if len(token) < 7 || token[:7] != "Bearer " {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
	}
	token = token[7:]

	// Validate the token
	claims, err := m.validator.ValidateToken(ctx, token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Extract custom claims
	customClaims, ok := claims.(*CustomClaims)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to extract custom claims")
	}

	// Add user information to the context
	ctx = context.WithValue(ctx, "user", customClaims)
	return handler(ctx, req)
}

// GetUserFromContext extracts user information from the context
func GetUserFromContext(ctx context.Context) (*CustomClaims, error) {
	claims, ok := ctx.Value("user").(*CustomClaims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}
	return claims, nil
}
