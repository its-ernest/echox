package auth

import (
    "net/http"
    "time"

    "otp-backend/internal/middleware"
    "otp-backend/internal/service"
    "otp-backend/internal/utils"

    "github.com/its-ernest/echox/store"
    "github.com/labstack/echo/v5"
    "github.com/golang-jwt/jwt/v5"
)

type Handler struct {
    service *service.AuthService
    jwtSecret []byte
}

func NewHandler(s *service.AuthService, secret []byte) *Handler {
    return &Handler{
        service: s,
        jwtSecret: secret,
    }
}

type requestOTP struct {
    Phone string `json:"phone"`
}

type verifyOTP struct {
    Phone string `json:"phone"`
    Code  string `json:"code"`
}

func (h *Handler) Register(g *echo.Group, store store.Store) {
	g.POST("/request", h.requestOTP, middleware.OTPCache(store))
	g.POST("/verify", h.verifyOTP)
}

func (h *Handler) requestOTP(c *echo.Context) error {
    req := new(requestOTP)
    if err := c.Bind(req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON")
    }

    // number validation
    if req.Phone == "" {
        return echo.NewHTTPError(http.StatusBadRequest, "Phone number is required")
    }
	// format for es64 format
    phone := utils.FormatPhone(req.Phone)

	// utilize auth_service.go's RequestOTP function
    if err := h.service.RequestOTP(c.Request().Context(), phone); err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "OTP failed")
    }

    return c.JSON(http.StatusOK, map[string]string{"message": "OTP sent"})
}

func (h *Handler) verifyOTP(c *echo.Context) error {
	// validate JSON request
    req := new(verifyOTP)
    if err := c.Bind(req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON")
    }
	
	// verify OTP stored in echox cache
    phone := utils.FormatPhone(req.Phone)
    if err := h.service.VerifyOTP(c.Request().Context(), phone, req.Code); err != nil {
        return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
    }

    // create and hand over a jwt afterwards
    claims := &jwt.RegisteredClaims{
        Subject:   phone,
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
    }
	// sign jwt
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    t, err := token.SignedString(h.jwtSecret)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "Could not generate token")
    }

	// deliver jwt
    return c.JSON(http.StatusOK, map[string]string{
        "message": "Verified",
        "token":   t,
    })
}