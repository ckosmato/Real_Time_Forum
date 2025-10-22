package handlers

import (
	"encoding/json"
	"net/http"
	"real-time-forum/models"
	"real-time-forum/services"
	"strings"
)

type AuthHandler struct {
	authService    services.AuthService
	sessionService services.SessionService
}

func NewAuthHandler(as services.AuthService, ss services.SessionService) *AuthHandler {
	return &AuthHandler{
		authService:    as,
		sessionService: ss,
	}
}

// -------- Helper functions --------





func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var input models.User
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			
			return
		}

		input.Email = strings.TrimSpace(input.Email)
		input.Nickname = strings.TrimSpace(input.Nickname)
		input.Password = strings.TrimSpace(input.Password)

		if input.Email == "" || input.Nickname == "" || input.Password == "" {
			//go standard
			http.Error(w, "All fields are required", http.StatusBadRequest)
			return
		}

		err := h.authService.RegisterUser(r.Context(), &input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Registration successful! Please log in.",
			})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var input models.User
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		if strings.Contains(input.Nickname, "@") {
			input.Email = input.Nickname
			input.Nickname = ""
		}

		user, err := h.authService.LoginUser(r.Context(), &input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		session, err := h.sessionService.GenerateSession(r.Context(), user)
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		// Set cookies
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    session.ID,
			Path:     "/",
			HttpOnly: true,
			Secure:   false, // change to true if using HTTPS
		})

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message":    "Login successful",
				"user":       user.Nickname,
				"session_id": session.ID,
			})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}


func (h *AuthHandler) LogOut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err == nil {
		sessionID := cookie.Value
		if err := h.sessionService.ExpireSession(r.Context(), sessionID); err != nil {
			http.Error(w, "Failed to log out", http.StatusInternalServerError)
			return
		}
	}

	// Expire session cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logout successful"})
}
