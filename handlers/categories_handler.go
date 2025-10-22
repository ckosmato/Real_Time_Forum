package handlers

import (
	"log"
	"net/http"
	"real-time-forum/services"
)

type CategoriesHandler struct {
	categoriesService services.CategoriesService
}

func NewCategoriesHandler(cs services.CategoriesService) *CategoriesHandler {
	return &CategoriesHandler{
		categoriesService: cs,
	}
}

func (h *CategoriesHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		log.Printf("CreateCategory: invalid method %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	categoryName := r.FormValue("category_name")

	err := h.categoriesService.CreateCategory(r.Context(), categoryName)
	if err != nil {
		log.Printf("CreateCategory: failed to create category: %v", err)
		// Set flash cookie if you have a flash cookie implementation
		// utils.SetFlashCookie(w, "Failed to create Category: "+err.Error())
		return
	}
}
