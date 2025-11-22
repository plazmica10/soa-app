package grpc

import (
	"context"

	pb "github.com/IvanNovakovic/SOA_Proj/protos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"follower-service/repository"
)

type FollowerServer struct {
	pb.UnimplementedFollowerServiceServer
	repo *repository.NeoRepository
}

func NewFollowerServer(repo *repository.NeoRepository) *FollowerServer {
	return &FollowerServer{repo: repo}
}

func (s *FollowerServer) GetFollowers(ctx context.Context, req *pb.GetFollowersRequest) (*pb.GetFollowersResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}

	followers, err := s.repo.Followers(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get followers: "+err.Error())
	}

	var pbFollowers []*pb.Follower
	for _, f := range followers {
		pbFollowers = append(pbFollowers, &pb.Follower{
			UserId: f,
		})
	}

	return &pb.GetFollowersResponse{
		Followers: pbFollowers,
	}, nil
}
