package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"real-time-forum/services"
	"real-time-forum/utils"
	"strings"
)

type DashboardHandler struct {
	postService       services.PostService
	categoriesService services.CategoriesService
	userService       services.UserService
}

func NewDashboardHandler(ps services.PostService, cs services.CategoriesService, us services.UserService) *DashboardHandler {
	return &DashboardHandler{
		postService:       ps,
		categoriesService: cs,
		userService:       us,
	}
}

func (h *DashboardHandler) Home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("Home: invalid method %s", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// Accept both "/" and "/dashboard" paths
	if r.URL.Path != "/" && r.URL.Path != "/dashboard" {
		log.Printf("Home: invalid path %s", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Page not found"})
		return
	}

	posts, err := h.postService.GetAllPosts(r.Context())
	if err != nil {
		log.Printf("Home: failed to fetch posts: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch posts"})
		return
	}

	// Try to get user from session, but don't fail if not found
	user := utils.GetUserFromContext(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":       user,
		"posts":      posts,
	})
}

func (h *DashboardHandler) PostsByCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("PostsByCategory: invalid method %s", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	categories, err := h.categoriesService.GetAllCategories(r.Context())
	if err != nil {
		log.Printf("PostsByCategory: failed to fetch categories: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch categories"})
		return
	}

	const prefix = "/category/"
	path := r.URL.Path
	if !strings.HasPrefix(path, prefix) {
		log.Printf("PostsByCategory: path doesn't have category prefix: %s", path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Page not found"})
		return
	}
	categoryID := strings.TrimPrefix(path, prefix)
	if categoryID == "" {
		log.Printf("PostsByCategory: empty category ID")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid category ID"})
		return
	}

	posts, err := h.postService.GetPostsByCategory(r.Context(), categoryID)
	if err != nil {
		log.Printf("PostsByCategory: failed to fetch posts for category %s: %v", categoryID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch posts"})
		return
	}

	// Try to get user from session header, but don't fail if not found
	user := utils.GetUserFromContext(r.Context())


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":       user,
		"posts":      posts,
		"categories": categories,
	})
}

func (h *DashboardHandler) UserPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("UserPosts: invalid method %s", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	user := utils.GetUserFromContext(r.Context())

	posts, err := h.postService.GetUserPosts(r.Context(), user.ID)
	if err != nil {
		log.Printf("UserPosts: failed to fetch user posts: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch user posts"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts":      posts,
	})
}
