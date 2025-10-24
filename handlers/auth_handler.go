package handlers

import (
	"encoding/json"
	"net/http"
	"real-time-forum/models"
	"real-time-forum/services"
	"regexp"
	"strconv"
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
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var user models.User

	// Handle form data request. Support both urlencoded and multipart (FormData) bodies.
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		// Allow up to 32MB in memory before spooling to disk
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid multipart form data"})
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid form data"})
			return
		}
	}

	// Convert form values to user struct
	ageStr := strings.TrimSpace(r.FormValue("age"))
	if ageStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Age is required"})
		return
	}

	age, err := strconv.Atoi(ageStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid age format"})
		return
	}

	user = models.User{
		Nickname:  strings.TrimSpace(r.FormValue("nickname")),
		FirstName: strings.TrimSpace(r.FormValue("firstName")),
		LastName:  strings.TrimSpace(r.FormValue("lastName")),
		Email:     strings.TrimSpace(r.FormValue("email")),
		Age:       age,
		Gender:    strings.TrimSpace(r.FormValue("gender")),
		Password:  r.FormValue("password"),
	}

	// Validate required fields
	if user.Nickname == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Nickname is required"})
		return
	}
	if user.FirstName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "First name is required"})
		return
	}
	if user.LastName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Last name is required"})
		return
	}
	if user.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email is required"})
		return
	}
	if user.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Password is required"})
		return
	}
	if user.Age <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Age is required"})
		return
	}
	if user.Gender == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Gender is required"})
		return
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(user.Email) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid email format"})
		return
	}

	// Validate age range
	if user.Age < 13 || user.Age > 120 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Age must be between 13 and 120"})
		return
	}

	// Validate password length
	if len(user.Password) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Password must be at least 6 characters long"})
		return
	}

	// Validate gender
	validGenders := map[string]bool{"Male": true, "Female": true, "Other": true}
	if !validGenders[user.Gender] {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid gender selection"})
		return
	}

	// Attempt to register user
	if err := h.authService.Register(r.Context(), &user); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			w.WriteHeader(http.StatusConflict)
			if strings.Contains(err.Error(), "email") {
				json.NewEncoder(w).Encode(map[string]string{"error": "Email already exists"})
			} else if strings.Contains(err.Error(), "nickname") {
				json.NewEncoder(w).Encode(map[string]string{"error": "Nickname already exists"})
			} else {
				json.NewEncoder(w).Encode(map[string]string{"error": "User already exists"})
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Registration failed: " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var input models.User
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			// JSON body
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON body"})
				return
			}
		} else {
			// Try form values (supports application/x-www-form-urlencoded and multipart/form-data)
			if strings.Contains(contentType, "multipart/form-data") {
				if err := r.ParseMultipartForm(32 << 20); err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{"error": "Invalid multipart form data"})
					return
				}
			} else {
				if err := r.ParseForm(); err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{"error": "Invalid form data"})
					return
				}
			}
			input.Nickname = strings.TrimSpace(r.FormValue("nickname"))
			input.Email = strings.TrimSpace(r.FormValue("email"))
			input.Password = r.FormValue("password")

			if strings.Contains(input.Nickname, "@") {
				input.Email = input.Nickname
				input.Nickname = ""
			}
		}

		user, err := h.authService.LoginUser(r.Context(), &input)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		session, err := h.sessionService.GenerateSession(r.Context(), user)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create session"})
			return
		}

		// Set cookies with more permissive settings for development
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    session.ID,
			Path:     "/",
			HttpOnly: false, // Allow JavaScript access for development
			Secure:   false, // Allow HTTP for development
			SameSite: http.SameSiteLaxMode,
			Domain:   "", // Browser will automatically set to current domain
			MaxAge:   3600, // 1 hour
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":    "Login successful",
			"user":       user.Nickname,
			"session_id": session.ID,
		})

	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
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
