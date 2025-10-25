package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"real-time-forum/handlers"
	"real-time-forum/repositories"
	"real-time-forum/services"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // Add this import
)

type Dependencies struct {
	AuthService    services.AuthService
	UserService    services.UserService
	SessionService services.SessionService

	PostService       services.PostService
	CategoriesService services.CategoriesService
	CommentService    services.CommentsService
}

type Handlers struct {
	AuthHandler       *handlers.AuthHandler
	DashboardHandler  *handlers.DashboardHandler
	PostHandler       *handlers.PostHandler
	CommentsHandler   *handlers.CommentsHandler
}

func main() {
	// Initialize database
	db, err := sql.Open("sqlite3", "./database/forum.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Setup dependencies and handlers
	deps := SetupDependencies(db)
	handlers := SetupHandlers(deps)

	// Setup routes
	mux := http.NewServeMux()
	Configure(mux, handlers)

	// Start background tasks
	go BackgroundTasks(deps.SessionService, deps.UserService)

	port := ":8080"
	println("Server listening on", port)
	println("Open http://localhost:8080 in your browser to view the forum")
	if err := http.ListenAndServe(port, mux); err != nil {
		panic(err)
	}

}

func Configure(mux *http.ServeMux, h *Handlers) {
	// Add logging middleware
	loggingHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("%s %s %s\n", r.Method, r.URL.Path, r.Proto)
			next.ServeHTTP(w, r)
		})
	}

	// API routes first (more specific patterns)
	mux.Handle("/register", loggingHandler(http.HandlerFunc(h.AuthHandler.Register)))
	mux.Handle("/login", loggingHandler(http.HandlerFunc(h.AuthHandler.Login)))
	mux.Handle("/logout", loggingHandler(http.HandlerFunc(h.AuthHandler.LogOut)))
	mux.Handle("/dashboard", loggingHandler(http.HandlerFunc(h.DashboardHandler.Home)))
	mux.Handle("/dashboard/my-posts", loggingHandler(http.HandlerFunc(h.DashboardHandler.UserPosts)))
	mux.Handle("/dashboard/active-users", loggingHandler(http.HandlerFunc(h.DashboardHandler.ActiveUsers)))
	mux.Handle("/createpost", loggingHandler(http.HandlerFunc(h.PostHandler.CreatePost)))
	mux.Handle("/post", loggingHandler(http.HandlerFunc(h.PostHandler.ViewPost)))
	mux.Handle("/post/createcomment", loggingHandler(http.HandlerFunc(h.CommentsHandler.CreateComment)))
	mux.Handle("/category/", loggingHandler(http.HandlerFunc(h.DashboardHandler.PostsByCategory)))

	// Static files
	mux.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "style.css")
	})
	mux.HandleFunc("/app.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "app.js")
	})

	// Root handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Root handler: %s %s %s\n", r.Method, r.URL.Path, r.Header.Get("Accept"))

		// If this is a request for JSON data (either through Accept header or query param)
		wantsJSON := strings.Contains(r.Header.Get("Accept"), "application/json") ||
			r.URL.Query().Get("format") == "json"

		if r.URL.Path == "/" && wantsJSON {
			h.DashboardHandler.Home(w, r)
			return
		}

		// Serve SPA for HTML requests or when no specific format is requested
		http.ServeFile(w, r, "index.html")
	})
}

func SetupDependencies(db *sql.DB) *Dependencies {
	// Repositories
	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	postRepo := repositories.NewPostRepository(db)
	categoriesRepo := repositories.NewCategoriesRepository(db)
	commentRepo := repositories.NewCommentRepository(db)

	// Services
	userService := services.NewUserService(*userRepo)
	authService := services.NewAuthService(*userRepo)

	sessionService := services.NewSessionService(*sessionRepo)
	postService := services.NewPostService(*postRepo)
	categoriesService := services.NewCategoriesService(*categoriesRepo)
	commentService := services.NewCommentsService(*commentRepo)

	return &Dependencies{
		UserService:       *userService,
		AuthService:       *authService,
		SessionService:    *sessionService,
		PostService:       *postService,
		CategoriesService: *categoriesService,
		CommentService:    *commentService,
	}
}

func SetupHandlers(deps *Dependencies) *Handlers {
	// Handlers
	return &Handlers{
		AuthHandler:       handlers.NewAuthHandler(deps.AuthService, deps.SessionService),
		CommentsHandler:   handlers.NewCommentsHandler(deps.PostService, deps.CommentService, deps.CategoriesService, deps.UserService),
		DashboardHandler:  handlers.NewDashboardHandler(deps.PostService, deps.CategoriesService, deps.UserService),
		PostHandler:       handlers.NewPostHandler(deps.PostService, deps.CategoriesService, deps.CommentService, deps.UserService),
	}
}


func BackgroundTasks(sessionService services.SessionService, userService services.UserService) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := sessionService.CleanupExpiredSessions(context.Background()); err != nil {
			log.Printf("Session cleanup error: %v", err)
		}
	}
}