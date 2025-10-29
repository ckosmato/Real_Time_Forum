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
	userService       services.UserService
}

func NewCommentsHandler(ps services.PostService, coms services.CommentsService, cs services.CategoriesService, us services.UserService) *CommentsHandler {
	return &CommentsHandler{
		postService:       ps,
		commentService:    coms,
		categoriesService: cs,
		userService:       us,
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

	// Get user from context
	user := utils.GetUserFromContext(r.Context())
	if user == nil {
		log.Printf("CreateComment: no user found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not logged in"})
		return
	}


	log.Printf("CreateComment: authenticated user ID='%s', Nickname='%s'", user.ID, user.Nickname)

	// Parse multipart form data
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		// If multipart parsing fails, try regular form parsing
		if err := r.ParseForm(); err != nil {
			log.Printf("CreateComment: invalid form data: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid form data"})
			return
		}
	}

	comment_input := r.FormValue("comment")
	postIDStr := r.FormValue("post_id")

	log.Printf("CreateComment: received comment='%s', post_id='%s'", comment_input, postIDStr)

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

	log.Printf("CreateComment: creating comment with AuthorID='%s', PostID='%s', Content='%s'", user.ID, postIDStr, comment_input)

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
