package main

import (
	"database/sql"
	"net/http"
	"real-time-forum/handlers"
	"real-time-forum/repositories"
	"real-time-forum/services"
)


type Dependencies struct {
	AuthService       services.AuthService
	UserService       services.UserService
	SessionService    services.SessionService

	PostService       services.PostService
	CategoriesService services.CategoriesService
	CommentService    services.CommentsService
}

type Handlers struct {
	AuthHandler       *handlers.AuthHandler
	DashboardHandler  *handlers.DashboardHandler
	PostHandler       *handlers.PostHandler
	CategoriesHandler *handlers.CategoriesHandler
	CommentsHandler   *handlers.CommentsHandler

}
func main() {
	

	port := ":8080"
	println("Server listening on", port)
    println("Open http://localhost:8080 in your browser to view the game")
	if err := http.ListenAndServe(port, nil); err != nil {
		panic(err)
	}
}

func Configure(mux *http.ServeMux,h *Handlers) {


	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/static"))))
	mux.Handle("/favicon.ico", http.RedirectHandler("/static/favicon/favicon.ico", http.StatusMovedPermanently))
	mux.Handle("/.well-known/", http.StripPrefix("/.well-known/", http.FileServer(http.Dir("static/.well-known/"))))

	//Dashboard Handler
	mux.HandleFunc("/",h.DashboardHandler.Home)
	mux.HandleFunc("/category/", h.DashboardHandler.PostsByCategory)


	// Auth Handler
	mux.HandleFunc("/register" , h.AuthHandler.Register)
	mux.HandleFunc("/login", h.AuthHandler.Login)
	mux.HandleFunc("/logout", h.AuthHandler.LogOut)




	// Post Handler
	mux.HandleFunc("/createpost", h.PostHandler.CreatePost)

	// Comments Handler
	mux.HandleFunc("/post", h.PostHandler.ViewPost)
	mux.HandleFunc("/post/createcomment", h.CommentsHandler.CreateComment)


	// Categories Handler
	mux.HandleFunc("/admin/createcategory", h.CategoriesHandler.CreateCategory)

}

func SetupDependencies(db *sql.DB) *Dependencies {
	// Repositories
	userRepo := repositories.NewUserRepository(db)
	sessionRepo :=repositories.NewSessionRepository(db)
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

func SetupHandlers(deps *Dependencies,) *Handlers {
	// Handlers
	return &Handlers{
		AuthHandler:       handlers.NewAuthHandler(deps.AuthService, deps.SessionService),
		CategoriesHandler: handlers.NewCategoriesHandler(deps.CategoriesService),
		CommentsHandler:   handlers.NewCommentsHandler(deps.PostService, deps.CommentService, deps.CategoriesService),
		DashboardHandler:  handlers.NewDashboardHandler(deps.PostService, deps.CategoriesService),
		PostHandler:       handlers.NewPostHandler(deps.PostService, deps.CategoriesService, deps.CommentService),
	}
}