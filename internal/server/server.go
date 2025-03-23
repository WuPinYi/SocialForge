package server

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/WuPinYi/SocialForge/internal/auth"
	"github.com/WuPinYi/SocialForge/internal/ent"
	"github.com/WuPinYi/SocialForge/internal/ent/influencer"
	"github.com/WuPinYi/SocialForge/internal/ent/post"
	"github.com/WuPinYi/SocialForge/internal/ent/user"
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

// User Management
func (s *Server) GetUser(ctx context.Context, req *ocsv1.GetUserRequest) (*ocsv1.GetUserResponse, error) {
	// Get the authenticated user's claims
	claims, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get the user from the database
	u, err := s.client.User.Get(ctx, req.Id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	// Check if the user has permission to view this user
	if u.Auth0ID != claims.Subject && claims.Subject != "admin" {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	return &ocsv1.GetUserResponse{
		User: &ocsv1.User{
			Id:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			Role:      u.Role,
			CreatedAt: timestamppb.New(u.CreatedAt),
			UpdatedAt: timestamppb.New(u.UpdatedAt),
		},
	}, nil
}

func (s *Server) ListUsers(ctx context.Context, req *ocsv1.ListUsersRequest) (*ocsv1.ListUsersResponse, error) {
	// Get the authenticated user's claims
	claims, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Only admin can list all users
	if claims.Subject != "admin" {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	query := s.client.User.Query()

	// Apply pagination
	if req.PageSize > 0 {
		query = query.Limit(int(req.PageSize))
	}
	if req.PageToken != "" {
		query = query.Where(user.IDGT(req.PageToken))
	}

	users, err := query.All(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	protoUsers := make([]*ocsv1.User, len(users))
	for i, u := range users {
		protoUsers[i] = &ocsv1.User{
			Id:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			Role:      u.Role,
			CreatedAt: timestamppb.New(u.CreatedAt),
			UpdatedAt: timestamppb.New(u.UpdatedAt),
		}
	}

	return &ocsv1.ListUsersResponse{
		Users: protoUsers,
		NextPageToken: func() string {
			if len(users) > 0 {
				return users[len(users)-1].ID
			}
			return ""
		}(),
	}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *ocsv1.UpdateUserRequest) (*ocsv1.UpdateUserResponse, error) {
	// Get the authenticated user's claims
	claims, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get the user to update
	u, err := s.client.User.Get(ctx, req.Id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	// Check if the user has permission to update this user
	if u.Auth0ID != claims.Subject && claims.Subject != "admin" {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	// Only admin can update roles
	if req.Role != "" && claims.Subject != "admin" {
		return nil, status.Error(codes.PermissionDenied, "only admin can update roles")
	}

	// Update the user
	update := s.client.User.UpdateOne(u)
	if req.Name != "" {
		update.SetName(req.Name)
	}
	if req.Role != "" {
		update.SetRole(req.Role)
	}

	u, err = update.Save(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &ocsv1.UpdateUserResponse{
		User: &ocsv1.User{
			Id:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			Role:      u.Role,
			CreatedAt: timestamppb.New(u.CreatedAt),
			UpdatedAt: timestamppb.New(u.UpdatedAt),
		},
	}, nil
}

// Influencer Management
func (s *Server) CreateInfluencer(ctx context.Context, req *ocsv1.CreateInfluencerRequest) (*ocsv1.CreateInfluencerResponse, error) {
	// Get the authenticated user's claims
	claims, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get the user
	u, err := s.client.User.Query().Where(user.Auth0IDEQ(claims.Subject)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// Create the user if they don't exist
			u, err = s.client.User.Create().
				SetID(uuid.New().String()).
				SetEmail(claims.Email).
				SetName(claims.Name).
				SetAuth0ID(claims.Subject).
				SetRole("user").
				Save(ctx)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
			}
		} else {
			return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
		}
	}

	// Create the influencer
	influencer, err := s.client.Influencer.Create().
		SetID(uuid.New().String()).
		SetName(req.Name).
		SetPlatform(req.Platform).
		SetAccountID(req.AccountId).
		SetOwner(u).
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
			OwnerId:   u.ID,
			CreatedAt: timestamppb.New(influencer.CreatedAt),
			UpdatedAt: timestamppb.New(influencer.UpdatedAt),
		},
	}, nil
}

func (s *Server) GetInfluencer(ctx context.Context, req *ocsv1.GetInfluencerRequest) (*ocsv1.GetInfluencerResponse, error) {
	// Get the authenticated user's claims
	claims, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get the influencer
	influencer, err := s.client.Influencer.Get(ctx, req.Id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "influencer not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get influencer: %v", err)
	}

	// Check if the user has permission to view this influencer
	if influencer.Edges.Owner.Auth0ID != claims.Subject && claims.Subject != "admin" {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	return &ocsv1.GetInfluencerResponse{
		Influencer: &ocsv1.Influencer{
			Id:        influencer.ID,
			Name:      influencer.Name,
			Platform:  influencer.Platform,
			AccountId: influencer.AccountID,
			Status:    influencer.Status,
			OwnerId:   influencer.Edges.Owner.ID,
			CreatedAt: timestamppb.New(influencer.CreatedAt),
			UpdatedAt: timestamppb.New(influencer.UpdatedAt),
		},
	}, nil
}

func (s *Server) ListInfluencers(ctx context.Context, req *ocsv1.ListInfluencersRequest) (*ocsv1.ListInfluencersResponse, error) {
	// Get the authenticated user's claims
	claims, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get the user
	u, err := s.client.User.Query().Where(user.Auth0IDEQ(claims.Subject)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	// Build the query
	query := s.client.Influencer.Query().Where(influencer.HasOwnerWith(user.ID(u.ID)))

	// Apply pagination
	if req.PageSize > 0 {
		query = query.Limit(int(req.PageSize))
	}
	if req.PageToken != "" {
		query = query.Where(influencer.IDGT(req.PageToken))
	}

	influencers, err := query.WithOwner().All(ctx)
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
			OwnerId:   inf.Edges.Owner.ID,
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

// Post Management
func (s *Server) SchedulePost(ctx context.Context, req *ocsv1.SchedulePostRequest) (*ocsv1.SchedulePostResponse, error) {
	// Get the authenticated user's claims
	claims, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get the influencer
	influencer, err := s.client.Influencer.Get(ctx, req.InfluencerId)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "influencer not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get influencer: %v", err)
	}

	// Check if the user has permission to create posts for this influencer
	if influencer.Edges.Owner.Auth0ID != claims.Subject && claims.Subject != "admin" {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	// Create the post
	post, err := s.client.Post.Create().
		SetID(uuid.New().String()).
		SetContent(req.Content).
		SetScheduledTime(req.ScheduledTime.AsTime()).
		SetInfluencer(influencer).
		Save(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create post: %v", err)
	}

	return &ocsv1.SchedulePostResponse{
		Post: &ocsv1.Post{
			Id:            post.ID,
			InfluencerId:  post.Edges.Influencer.ID,
			Content:       post.Content,
			ScheduledTime: timestamppb.New(post.ScheduledTime),
			Status:        post.Status,
			CreatedAt:     timestamppb.New(post.CreatedAt),
			UpdatedAt:     timestamppb.New(post.UpdatedAt),
		},
	}, nil
}

func (s *Server) GetPost(ctx context.Context, req *ocsv1.GetPostRequest) (*ocsv1.GetPostResponse, error) {
	// Get the authenticated user's claims
	claims, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get the post
	post, err := s.client.Post.Get(ctx, req.Id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "post not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get post: %v", err)
	}

	// Check if the user has permission to view this post
	if post.Edges.Influencer.Edges.Owner.Auth0ID != claims.Subject && claims.Subject != "admin" {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	return &ocsv1.GetPostResponse{
		Post: &ocsv1.Post{
			Id:            post.ID,
			InfluencerId:  post.Edges.Influencer.ID,
			Content:       post.Content,
			ScheduledTime: timestamppb.New(post.ScheduledTime),
			Status:        post.Status,
			CreatedAt:     timestamppb.New(post.CreatedAt),
			UpdatedAt:     timestamppb.New(post.UpdatedAt),
		},
	}, nil
}

func (s *Server) ListPosts(ctx context.Context, req *ocsv1.ListPostsRequest) (*ocsv1.ListPostsResponse, error) {
	// Get the authenticated user's claims
	claims, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get the influencer
	influencer, err := s.client.Influencer.Get(ctx, req.InfluencerId)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "influencer not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get influencer: %v", err)
	}

	// Check if the user has permission to view posts for this influencer
	if influencer.Edges.Owner.Auth0ID != claims.Subject && claims.Subject != "admin" {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	// Build the query
	query := s.client.Post.Query().Where(post.InfluencerID(influencer.ID))

	// Apply pagination
	if req.PageSize > 0 {
		query = query.Limit(int(req.PageSize))
	}
	if req.PageToken != "" {
		query = query.Where(post.IDGT(req.PageToken))
	}

	posts, err := query.WithInfluencer().All(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list posts: %v", err)
	}

	protoPosts := make([]*ocsv1.Post, len(posts))
	for i, p := range posts {
		protoPosts[i] = &ocsv1.Post{
			Id:            p.ID,
			InfluencerId:  p.Edges.Influencer.ID,
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
