package services

import (
	"context"
	"errors"
	"log"
	"real-time-forum/models"
	repositories "real-time-forum/repositories"
	"strings"

	"github.com/gofrs/uuid"
)

type CommentsService struct {
	repo repositories.CommentRepository
}

func NewCommentsService(repo repositories.CommentRepository) *CommentsService {
	return &CommentsService{repo: repo}
}

func (s *CommentsService) CreateComment(ctx context.Context, comment *models.Comment) error {
	if strings.TrimSpace(comment.Content) == "" {
		log.Printf("CreateComment: comment cannot be empty")
		return errors.New("comment cannot be empty")
	}
	u1, err := uuid.NewV4()
	if err != nil {
		log.Printf("CreateComment: failed to generate comment ID: %v", err)
		return errors.New("failed to generate comment ID")
	}

	comment.ID = u1.String()
	if err := s.repo.CreateComment(ctx, comment); err != nil {
		log.Printf("CreateComment: failed to create comment: %v", err)
		return errors.New("failed to create comment")
	}
	return nil
}

func (s *CommentsService) GetPostComments(ctx context.Context, postID string) ([]models.Comment, error) {
	postComments, err := s.repo.GetPostComments(ctx, postID)
	if err != nil {
		log.Printf("GetPostComments: failed to retrieve comments for post %s: %v", postID, err)
		return nil, errors.New("failed to retrieve comments")
	}
	return postComments, nil
}

func (s *CommentsService) GetCommentByID(ctx context.Context, commentID string) (*models.Comment, error) {
	comment, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		log.Printf("GetCommentByID: failed to retrieve comment %s: %v", commentID, err)
		return nil, errors.New("failed to retrieve comment")
	}
	return comment, nil
}
