package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"real-time-forum/models"
	"real-time-forum/services"
)

type PostHandler struct {
	postService       services.PostService
	categoriesService services.CategoriesService
	commentService    services.CommentsService
	userService       services.UserService
}

func NewPostHandler(ps services.PostService, cs services.CategoriesService, coms services.CommentsService, us services.UserService) *PostHandler {
	return &PostHandler{
		postService:       ps,
		categoriesService: cs,
		commentService:    coms,
		userService:       us,
	}
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		categories, err := h.categoriesService.GetAllCategories(r.Context())
		if err != nil {
			log.Printf("CreatePost GET: failed to fetch categories: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch categories"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"categories": categories,
		})

	case http.MethodPost:
		// Get user from session cookie
		sessionCookie, err := r.Cookie("session_id")
		if err != nil {
			log.Printf("CreatePost: no session cookie found: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Not logged in"})
			return
		}

		// Validate session and get user
		user, err := h.userService.GetUserBySessionID(r.Context(), sessionCookie.Value)
		if err != nil {
			log.Printf("CreatePost: invalid session: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid session"})
			return
		}

		log.Printf("CreatePost: authenticated user ID=%s, nickname=%s", user.ID, user.Nickname)

		if err := r.ParseMultipartForm(20 << 20); err != nil {
			log.Printf("CreatePost: invalid form: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid form"})
			return
		}

		post := models.Post{
			Title:   r.FormValue("title"),
			Content: r.FormValue("content"),
		}

		var catIDs []string
		catIDs = append(catIDs, r.Form["categories"]...)

		if len(catIDs) == 0 {
			log.Printf("CreatePost: no categories selected")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Pick at least one category"})
			return
		}

		if err := h.postService.CreatePost(r.Context(), user, &post, catIDs); err != nil {
			log.Printf("CreatePost: failed to create post: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create post"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Post created successfully",
			"post":    post,
		})

	default:
		log.Printf("CreatePost: invalid method %s", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}

func (h *PostHandler) ViewPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("ViewPost: invalid method %s", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// Get user from session cookie
	var user *models.User
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("ViewPost: no session cookie found: %v", err)
		// Allow viewing posts without being logged in, but user will be nil
		user = nil
	} else {
		// Validate session and get user
		user, err = h.userService.GetUserBySessionID(r.Context(), sessionCookie.Value)
		if err != nil {
			log.Printf("ViewPost: invalid session: %v", err)
			user = nil // Don't fail, just set user to nil
		}
	}

	query := r.URL.Query()
	postIDStr := query.Get("id")
	log.Printf("ViewPost: received post ID: '%s'", postIDStr)
	if postIDStr == "" {
		log.Printf("ViewPost: missing post ID")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Post ID is required"})
		return
	}

	post, err := h.postService.GetPostByID(r.Context(), postIDStr)
	if err != nil {
		log.Printf("ViewPost: failed to fetch post %s: %v", postIDStr, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Post not found"})
		return
	}

	commentDisplay, err := h.commentService.GetPostComments(r.Context(), post.ID)
	if err != nil {
		log.Printf("ViewPost: failed to fetch comments for post %s: %v", post.ID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch comments"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":     user,
		"post":     post,
		"comments": commentDisplay,
	})
}
