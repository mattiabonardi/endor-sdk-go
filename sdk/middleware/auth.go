package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// AuthMiddleware provides authentication and authorization middleware.
// It validates session tokens and ensures proper security context propagation.
type AuthMiddleware struct {
	Config      interfaces.ConfigProviderInterface // For auth configuration
	Logger      interfaces.LoggerInterface         // For auth logging
	RequireAuth bool                               // Whether authentication is mandatory
}

// NewAuthMiddleware creates a new authentication middleware with dependency injection.
func NewAuthMiddleware(deps AuthMiddlewareDependencies) *AuthMiddleware {
	return &AuthMiddleware{
		Config:      deps.Config,
		Logger:      deps.Logger,
		RequireAuth: deps.RequireAuth,
	}
}

// AuthMiddlewareDependencies contains all required dependencies for AuthMiddleware.
type AuthMiddlewareDependencies struct {
	Config      interfaces.ConfigProviderInterface
	Logger      interfaces.LoggerInterface
	RequireAuth bool
}

// Before validates authentication and sets up security context.
func (a *AuthMiddleware) Before(ctx interface{}) error {
	middlewareCtx, ok := ctx.(*middlewareContext)
	if !ok {
		return fmt.Errorf("auth middleware requires middleware context, got %T", ctx)
	}

	// Type assert to get access to HTTP headers - this will be handled by the adapter
	// For now, we'll use reflection to get the headers safely
	headers := extractHeaders(middlewareCtx.ginContext)
	if headers == nil {
		return fmt.Errorf("auth middleware could not access request headers")
	}

	sessionToken := getHeader(headers, "x-user-session")
	userID := getHeader(headers, "x-user-id")

	// If no token and auth is required, reject request
	if sessionToken == "" && a.RequireAuth {
		a.Logger.Warn("Authentication required but no session token provided")
		return errors.New("authentication required: missing session token")
	}

	// Basic session token validation (simplified for MVP)
	if sessionToken != "" {
		// Simple validation - in production this would validate against a token store
		if len(sessionToken) < 10 {
			a.Logger.Warn("Invalid session token format", "token_length", len(sessionToken))
			return errors.New("authentication failed: invalid token format")
		}

		if userID == "" {
			a.Logger.Warn("Session token provided but no user ID")
			return errors.New("authentication failed: missing user identifier")
		}

		// Set additional context for downstream handlers (implementation specific)
		setContextValue(middlewareCtx.ginContext, "authenticated", true)
		setContextValue(middlewareCtx.ginContext, "userID", userID)
		setContextValue(middlewareCtx.ginContext, "sessionToken", sessionToken)

		a.Logger.Debug("Authentication successful", "userID", userID)
	} else {
		setContextValue(middlewareCtx.ginContext, "authenticated", false)
	}

	return nil
}

// After cleans up security context and logs authentication metrics.
func (a *AuthMiddleware) After(ctx interface{}, response interface{}) error {
	middlewareCtx, ok := ctx.(*middlewareContext)
	if !ok {
		return fmt.Errorf("auth middleware requires middleware context, got %T", ctx)
	}

	// Extract headers for logging
	headers := extractHeaders(middlewareCtx.ginContext)
	if headers == nil {
		return fmt.Errorf("auth middleware could not access request headers for logging")
	}

	sessionToken := getHeader(headers, "x-user-session")
	userID := getHeader(headers, "x-user-id")

	// Log authentication metrics
	authenticated := sessionToken != ""
	status := http.StatusOK // Default status, actual status may vary

	a.Logger.Info("Request completed",
		"authenticated", authenticated,
		"userID", userID,
		"status", status)

	return nil
}
