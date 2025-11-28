package grpc

import (
	"context"
	"log"

	"tour-service/model"
	"tour-service/repository"

	pb "github.com/IvanNovakovic/SOA_Proj/protos"
)

type TourGRPCServer struct {
	pb.UnimplementedTourServiceServer
	repo *repository.TourRepository
}

func NewTourGRPCServer(repo *repository.TourRepository) *TourGRPCServer {
	return &TourGRPCServer{repo: repo}
}

// GetTourByID implements the GetTourByID RPC method
func (s *TourGRPCServer) GetTourByID(ctx context.Context, req *pb.GetTourByIDRequest) (*pb.GetTourByIDResponse, error) {
	log.Printf("gRPC GetTourByID called with tour_id: %s", req.TourId)

	tour, err := s.repo.GetTourByID(ctx, req.TourId)
	if err != nil {
		log.Printf("Error fetching tour: %v", err)
		return nil, err
	}

	pbTour := convertTourToProto(tour)
	return &pb.GetTourByIDResponse{Tour: pbTour}, nil
}

// GetToursByAuthor implements the GetToursByAuthor RPC method
func (s *TourGRPCServer) GetToursByAuthor(ctx context.Context, req *pb.GetToursByAuthorRequest) (*pb.GetToursByAuthorResponse, error) {
	log.Printf("gRPC GetToursByAuthor called with author_id: %s", req.AuthorId)

	tours, err := s.repo.GetToursByAuthor(ctx, req.AuthorId)
	if err != nil {
		log.Printf("Error fetching tours: %v", err)
		return nil, err
	}

	var pbTours []*pb.Tour
	for _, tour := range tours {
		pbTours = append(pbTours, convertTourToProto(&tour))
	}

	return &pb.GetToursByAuthorResponse{Tours: pbTours}, nil
}

// Helper function to convert model.Tour to protobuf Tour
func convertTourToProto(tour *model.Tour) *pb.Tour {
	pbTour := &pb.Tour{
		Id:          tour.ID.Hex(),
		AuthorId:    tour.AuthorID,
		Name:        tour.Name,
		Description: tour.Description,
		Difficulty:  tour.Difficulty,
		Tags:        tour.Tags,
		Status:      tour.Status,
		Price:       tour.Price,
		Distance:    tour.Distance,
		Durations: &pb.TransportDuration{
			Walking: int32(tour.Durations.Walking),
			Biking:  int32(tour.Durations.Biking),
			Driving: int32(tour.Durations.Driving),
		},
		CreatedAt: tour.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if tour.PublishedAt != nil {
		pbTour.PublishedAt = tour.PublishedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	if tour.ArchivedAt != nil {
		pbTour.ArchivedAt = tour.ArchivedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return pbTour
}
