package server

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/WuPinYi/SocialForge/internal/ent"
	"github.com/WuPinYi/SocialForge/internal/ent/influencer"
	"github.com/WuPinYi/SocialForge/internal/ent/post"
	ocsv1 "github.com/WuPinYi/SocialForge/proto/ocs/v1"
)

type Server struct {
	ocsv1.UnimplementedOpinionControlServiceServer
	client *ent.Client
}

func NewServer(client *ent.Client) *Server {
	return &Server{
		client: client,
	}
}

// CreateInfluencer creates a new influencer
func (s *Server) CreateInfluencer(ctx context.Context, req *ocsv1.CreateInfluencerRequest) (*ocsv1.CreateInfluencerResponse, error) {
	influencer, err := s.client.Influencer.Create().
		SetID(uuid.New().String()).
		SetName(req.Name).
		SetPlatform(req.Platform).
		SetAccountID(req.AccountId).
		Save(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create influencer: %v", err)
	}

	return &ocsv1.CreateInfluencerResponse{
		Influencer: &ocsv1.Influencer{
			Id:        influencer.ID,
			Name:      influencer.Name,
			Platform:  influencer.Platform,
			AccountId: influencer.AccountID,
			Status:    influencer.Status,
			CreatedAt: timestamppb.New(influencer.CreatedAt),
			UpdatedAt: timestamppb.New(influencer.UpdatedAt),
		},
	}, nil
}

// GetInfluencer retrieves an influencer by ID
func (s *Server) GetInfluencer(ctx context.Context, req *ocsv1.GetInfluencerRequest) (*ocsv1.GetInfluencerResponse, error) {
	influencer, err := s.client.Influencer.Get(ctx, req.Id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "influencer not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get influencer: %v", err)
	}

	return &ocsv1.GetInfluencerResponse{
		Influencer: &ocsv1.Influencer{
			Id:        influencer.ID,
			Name:      influencer.Name,
			Platform:  influencer.Platform,
			AccountId: influencer.AccountID,
			Status:    influencer.Status,
			CreatedAt: timestamppb.New(influencer.CreatedAt),
			UpdatedAt: timestamppb.New(influencer.UpdatedAt),
		},
	}, nil
}

// ListInfluencers retrieves a list of influencers with pagination
func (s *Server) ListInfluencers(ctx context.Context, req *ocsv1.ListInfluencersRequest) (*ocsv1.ListInfluencersResponse, error) {
	query := s.client.Influencer.Query()

	// Apply pagination
	if req.PageSize > 0 {
		query = query.Limit(int(req.PageSize))
	}
	if req.PageToken != "" {
		// Implement cursor-based pagination here
		// This is a simplified version
		query = query.Where(influencer.IDGT(req.PageToken))
	}

	influencers, err := query.All(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list influencers: %v", err)
	}

	protoInfluencers := make([]*ocsv1.Influencer, len(influencers))
	for i, inf := range influencers {
		protoInfluencers[i] = &ocsv1.Influencer{
			Id:        inf.ID,
			Name:      inf.Name,
			Platform:  inf.Platform,
			AccountId: inf.AccountID,
			Status:    inf.Status,
			CreatedAt: timestamppb.New(inf.CreatedAt),
			UpdatedAt: timestamppb.New(inf.UpdatedAt),
		}
	}

	return &ocsv1.ListInfluencersResponse{
		Influencers: protoInfluencers,
		NextPageToken: func() string {
			if len(influencers) > 0 {
				return influencers[len(influencers)-1].ID
			}
			return ""
		}(),
	}, nil
}

// SchedulePost schedules a new post for an influencer
func (s *Server) SchedulePost(ctx context.Context, req *ocsv1.SchedulePostRequest) (*ocsv1.SchedulePostResponse, error) {
	// Verify influencer exists
	_, err := s.client.Influencer.Get(ctx, req.InfluencerId)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "influencer not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to verify influencer: %v", err)
	}

	post, err := s.client.Post.Create().
		SetID(uuid.New().String()).
		SetInfluencerID(req.InfluencerId).
		SetContent(req.Content).
		SetScheduledTime(req.ScheduledTime.AsTime()).
		Save(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to schedule post: %v", err)
	}

	return &ocsv1.SchedulePostResponse{
		Post: &ocsv1.Post{
			Id:            post.ID,
			InfluencerId:  post.InfluencerID,
			Content:       post.Content,
			ScheduledTime: timestamppb.New(post.ScheduledTime),
			Status:        post.Status,
			CreatedAt:     timestamppb.New(post.CreatedAt),
			UpdatedAt:     timestamppb.New(post.UpdatedAt),
		},
	}, nil
}

// GetPost retrieves a post by ID
func (s *Server) GetPost(ctx context.Context, req *ocsv1.GetPostRequest) (*ocsv1.GetPostResponse, error) {
	post, err := s.client.Post.Get(ctx, req.Id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "post not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get post: %v", err)
	}

	return &ocsv1.GetPostResponse{
		Post: &ocsv1.Post{
			Id:            post.ID,
			InfluencerId:  post.InfluencerID,
			Content:       post.Content,
			ScheduledTime: timestamppb.New(post.ScheduledTime),
			Status:        post.Status,
			CreatedAt:     timestamppb.New(post.CreatedAt),
			UpdatedAt:     timestamppb.New(post.UpdatedAt),
		},
	}, nil
}

// ListPosts retrieves a list of posts for an influencer with pagination
func (s *Server) ListPosts(ctx context.Context, req *ocsv1.ListPostsRequest) (*ocsv1.ListPostsResponse, error) {
	query := s.client.Post.Query().
		Where(post.InfluencerID(req.InfluencerId))

	// Apply pagination
	if req.PageSize > 0 {
		query = query.Limit(int(req.PageSize))
	}
	if req.PageToken != "" {
		// Implement cursor-based pagination here
		// This is a simplified version
		query = query.Where(post.IDGT(req.PageToken))
	}

	posts, err := query.All(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list posts: %v", err)
	}

	protoPosts := make([]*ocsv1.Post, len(posts))
	for i, p := range posts {
		protoPosts[i] = &ocsv1.Post{
			Id:            p.ID,
			InfluencerId:  p.InfluencerID,
			Content:       p.Content,
			ScheduledTime: timestamppb.New(p.ScheduledTime),
			Status:        p.Status,
			CreatedAt:     timestamppb.New(p.CreatedAt),
			UpdatedAt:     timestamppb.New(p.UpdatedAt),
		}
	}

	return &ocsv1.ListPostsResponse{
		Posts: protoPosts,
		NextPageToken: func() string {
			if len(posts) > 0 {
				return posts[len(posts)-1].ID
			}
			return ""
		}(),
	}, nil
}
