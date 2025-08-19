package api

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"net/http"
	"new-billing/internal/config"
	"new-billing/internal/models"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB  *sqlx.DB
	Cfg *config.Config
}

// @Summary      Аутентификация пользователя
// @Description  Принимает имя пользователя и пароль, возвращает JWT токен в случае успеха.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        credentials body object{username=string,password=string} true "Учетные данные для входа"
// @Success      200 {object} object{token=string} "Успешная аутентификация с JWT токеном"
// @Failure      400 {object} map[string]string "Некорректный запрос"
// @Failure      401 {object} map[string]string "Ошибка аутентификации"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	user := models.User{}
	if err := h.DB.Get(&user, "SELECT * FROM users WHERE username=$1", creds.Username); err != nil {
		http.Error(w, `{"error": "invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		http.Error(w, `{"error": "invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.RegisteredClaims{
		Subject:   strconv.Itoa(user.ID),
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		Audience:  jwt.ClaimStrings{string(user.Role)},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.Cfg.Auth.JWTSecret))
	if err != nil {
		log.Printf("Error creating token: %v", err)
		http.Error(w, `{"error": "could not create token"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
