package api

import (
	"crypto/rand"
	"errors"
	"fmt"
	"logger"
	"time"

	"server/config"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// InitSecret generates a random secret key and assigns it to the config.
func InitSecret() ([]byte, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to generate secret")
	}
	return secret, nil
}

// VerifyToken validates the JWT token using the secret key.
func VerifyToken(token string) (*jwt.Token, error) {
	tok, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing algorithm: %v, expected HS256", t.Method.Alg())
		}
		return config.Secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token parse error: %w", err)
	}
	if !tok.Valid {
		return nil, errors.New("invalid token")
	}

	if claims, ok := tok.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return nil, errors.New("token is expired")
			}
		} else {
			return nil, errors.New("invalid expiration claim in token")
		}
	} else {
		return nil, errors.New("invalid token claims")
	}

	return tok, nil
}

// HashPassword hashes the password using bcrypt.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CreateJWT generates a new JWT token with an expiration time of 24 hours.
func CreateJWT(username string) (string, int64, error) {
	exp := time.Now().Add(24 * time.Hour).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      exp,
	})

	tokenString, err := token.SignedString(config.Secret)
	if err != nil {
		return "", 0, err
	}
	return tokenString, exp, nil
}

// HandleLogin handles the login request by checking the credentials and generating a JWT token.
//
// @Summary Login
// @Description Authenticates a user and issues a JWT cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SigninRequest true "Credentials"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func HandleLogin(c fiber.Ctx) error {
	req := new(SigninRequest)
	if err := c.Bind().Body(req); err != nil {
		logger.Log.Warn().Err(err).Msg("Invalid login payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if req.Password == "" {
		logger.Log.Warn().Msg("Missing password in login")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password is required",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(config.Password), []byte(req.Password)); err != nil {
		logger.Log.Warn().Msg("Login failed: invalid password")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	if req.Username == "" {
		req.Username = "cookieguest"
		logger.Log.Debug().Msg("Username not provided, using default 'cookieguest'")
	}

	token, _, err := CreateJWT(req.Username)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to generate JWT")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "JWT generation error",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		MaxAge:   60 * 60 * 48, // 2 day
		HTTPOnly: true,
		SameSite: "Strict",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{})
}

// HandleVerify verifies JWT token validity from cookie.
//
// @Summary Verify token
// @Description Verifies if JWT cookie token is valid and not expired.
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]string
// @Router /auth/verify [get]
func HandleVerify(c fiber.Ctx) error {
	token := c.Cookies("token")
	if token == "" {
		logger.Log.Warn().Msg("JWT cookie missing")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "JWT token is required",
		})
	}
	jwtToken, err := VerifyToken(token)
	if err != nil {
		logger.Log.Warn().Err(err).Msg("JWT verification failed")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired JWT token",
		})
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid JWT token",
		})
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		username = "cookieguest"
	}

	return c.Status(fiber.StatusOK).JSON(AuthVerifyResponse{
		Username: username,
	})
}

// HandleLogout clears the JWT cookie.
func HandleLogout(c fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		MaxAge:   -1,
		HTTPOnly: true,
		SameSite: "Strict",
		Expires:  time.Unix(0, 0),
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{})
}

// CookieAuthMiddleware checks if the user has a valid JWT token in their cookies.
func CookieAuthMiddleware(c fiber.Ctx) error {
	token := c.Cookies("token")
	if token == "" {
		logger.Log.Warn().Msg("JWT cookie missing")
		return c.Redirect().To("/login")
	}
	if _, err := VerifyToken(token); err != nil {
		logger.Log.Warn().Err(err).Msg("JWT verification failed")
		return c.Redirect().To("/login")
	}
	return nil
}
