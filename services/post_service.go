package services

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"real-time-forum/models"
	repositories "real-time-forum/repositories"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

type PostService struct {
	repo repositories.PostRepository
}

func NewPostService(repo repositories.PostRepository) *PostService {
	return &PostService{repo: repo}
}

func (s *PostService) CreatePost(ctx context.Context, user *models.User, post *models.Post, cat []string) error {
	if strings.TrimSpace(post.Title) == "" || strings.TrimSpace(post.Content) == "" {
		log.Printf("CreatePost: post title and content cannot be empty")
		return errors.New("post title and content cannot be empty")
	}

	u1, err := uuid.NewV4()
	if err != nil {
		log.Printf("CreatePost: failed to generate post ID: %v", err)
		return errors.New("failed to generate post ID")
	}

	post.ID = u1.String()
	post.CreatedAt = time.Now()
	err = s.repo.CreatePost(ctx, user, post, cat)
	if err != nil {
		log.Printf("CreatePost: failed to create post: %v", err)
		return errors.New("failed to create post")
	}
	return nil
}

func (s *PostService) GetAllPosts(ctx context.Context) ([]models.Post, error) {
	postsView, err := s.repo.GetAllPosts(ctx)
	if err != nil {
		log.Printf("GetAllPosts: failed to fetch posts: %v", err)
		return nil, errors.New("failed to fetch posts")
	}
	return postsView, nil
}

func (s *PostService) GetPostByID(ctx context.Context, postID string) (*models.Post, error) {
	postView, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("GetPostByID: post with ID %s not found", postID)
			return nil, errors.New("post not found")
		}
		log.Printf("GetPostByID: failed to fetch post %s: %v", postID, err)
		return nil, errors.New("failed to fetch post")
	}
	return postView, nil
}

func (s *PostService) GetPostsByCategory(ctx context.Context, categoryID string) ([]models.Post, error) {
	posts, err := s.repo.GetPostsByCategory(ctx, categoryID)
	if err != nil {
		log.Printf("GetPostsByCategory: failed to fetch posts by category %s: %v", categoryID, err)
		return nil, errors.New("failed to fetch posts by this category")
	}
	return posts, nil
}

func (s *PostService) GetUserPosts(ctx context.Context, userID string) ([]models.Post, error) {
	
	posts, err := s.repo.GetUserPosts(ctx, userID)
	if err != nil {
		log.Printf("GetUserPosts: failed to fetch posts of user %s: %v", userID, err)
		return nil, errors.New("failed to fetch posts of this user")
	}
	return posts, nil
}
