package services

import (
	"context"
	"real-time-forum/models"
	"real-time-forum/repositories"
	"sort"
)

type ActivityService interface {
	ListAll(ctx context.Context, userID string, limit, offset int) ([]models.ActivityItem, error)
	ListPosts(ctx context.Context, userID string, limit, offset int) ([]models.ActivityItem, error)
	ListComments(ctx context.Context, userID string, limit, offset int) ([]models.ActivityItem, error)
	ListReactions(ctx context.Context, userID string, limit, offset int) ([]models.ActivityItem, error)
}

type activityServiceImpl struct {
	postRepo    *sqlite.PostRepository
	commentRepo *sqlite.CommentRepository
}

// ListReactions implements ActivityService.
func (s *activityServiceImpl) ListReactions(ctx context.Context, userID string, limit int, offset int) ([]models.ActivityItem, error) {
	panic("unimplemented")
}

func NewActivityService(postR *sqlite.PostRepository, commR *sqlite.CommentRepository) ActivityService {
	return &activityServiceImpl{postRepo: postR, commentRepo: commR}
}

func (s *activityServiceImpl) ListAll(ctx context.Context, userID string, limit, offset int) ([]models.ActivityItem, error) {
	var err error

	posts, err := s.postRepo.ListByAuthor(ctx, userID, limit*2, 0)
	if err != nil {
		return nil, err
	}

	comments, err := s.commentRepo.CommentsListByUser(ctx, userID, limit*2, 0)
	if err != nil {
		return nil, err
	}

	var all []models.ActivityItem

	for _, p := range posts {
		all = append(all, models.ActivityItem{
			Type:      "post",
			Timestamp: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
			UserID:    p.AuthorID,
			PostID:    p.ID,
			PostTitle: p.Title,
		})
	}

	for _, c := range comments {
		all = append(all, models.ActivityItem{
			Type:      "comment",
			Timestamp: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			UserID:    c.AuthorID, // Not sure if it needs
			PostID:    c.PostID,
			PostTitle: c.PostTitle,
			CommentID: c.ID,
			Comment:   c.Content,
		})
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Timestamp.After(all[j].Timestamp)
	})

	if offset >= len(all) {
		return []models.ActivityItem{}, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (s *activityServiceImpl) ListPosts(ctx context.Context, userID string, limit, offset int) ([]models.ActivityItem, error) {
	posts, err := s.postRepo.ListByAuthor(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	var postsItems []models.ActivityItem
	for _, p := range posts {
		postsItems = append(postsItems, models.ActivityItem{
			Type:      "post",
			Timestamp: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
			UserID:    p.AuthorID,
			PostID:    p.ID,
			PostTitle: p.Title,
		})
	}
	return postsItems, nil
}

func (s *activityServiceImpl) ListComments(ctx context.Context, userID string, limit, offset int) ([]models.ActivityItem, error) {
	comments, err := s.commentRepo.CommentsListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	var commentsItems []models.ActivityItem
	for _, c := range comments {
		commentsItems = append(commentsItems, models.ActivityItem{
			Type:      "comment",
			Timestamp: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			UserID:    c.AuthorID, // Not sure if it needs
			PostID:    c.PostID,
			PostTitle: c.PostTitle,
			CommentID: c.ID,
			Comment:   c.Content,
		})
	}
	return commentsItems, nil
}
