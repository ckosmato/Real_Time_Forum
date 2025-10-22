package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"real-time-forum/models"
	"real-time-forum/services"
    "real-time-forum/utils"
	"strings"
)

type CommentsHandler struct {
	postService       services.PostService
	commentService    services.CommentsService
	categoriesService services.CategoriesService
}

func NewCommentsHandler(ps services.PostService, coms services.CommentsService, cs services.CategoriesService) *CommentsHandler {
	return &CommentsHandler{
		postService:       ps,
		commentService:    coms,
		categoriesService: cs,
	}
}

func (h *CommentsHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("CreateComment: invalid method %s", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	user := utils.GetUserFromContext(r.Context())

	if err := r.ParseForm(); err != nil {
		log.Printf("CreateComment: invalid form data: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid form data"})
		return
	}

	comment_input := r.FormValue("comment")
	postIDStr := r.FormValue("post_id")

	if strings.TrimSpace(comment_input) == "" {
		log.Printf("CreateComment: comment cannot be empty")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Comment cannot be empty"})
		return
	}

	comment := models.Comment{
		PostID:   postIDStr,
		AuthorID: user.ID,
		Content:  comment_input,
	}

	if err := h.commentService.CreateComment(r.Context(), &comment); err != nil {
		log.Printf("CreateComment: failed to create comment: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create comment"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Comment created successfully",
		"comment": comment,
	})
}
