package service

import (
	"context"
	"errors"
	pb "github.com/zalgonoise/x/collide/pkg/api/pb/collide/v1"
	"github.com/zalgonoise/x/collide/repository/memory"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Repository interface {
	ListDistricts(ctx context.Context) ([]string, error)
	ListAllTracksByDistrict(ctx context.Context, district string) ([]string, error)
	ListDriftTracksByDistrict(ctx context.Context, district string) ([]string, error)
	GetAlternativesByDistrictAndTrack(ctx context.Context, district string, trackID string) ([]string, error)
	GetCollisionsByDistrictAndTrack(ctx context.Context, district string, trackID string) ([]string, error)
}

type Service struct {
	pb.UnimplementedCollideServiceServer

	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListDistricts(ctx context.Context, _ *pb.ListDistrictsRequest) (*pb.ListDistrictsResponse, error) {
	districts, err := s.repo.ListDistricts(ctx)

	switch {
	case errors.Is(err, memory.ErrNoDistricts):
		return nil, status.Error(codes.NotFound, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.ListDistrictsResponse{Districts: districts}, nil
}

func (s *Service) ListAllTracksByDistrict(ctx context.Context, req *pb.ListAllTracksByDistrictRequest) (*pb.ListAllTracksByDistrictResponse, error) {
	tracks, err := s.repo.ListAllTracksByDistrict(ctx, req.GetDistrict())

	switch {
	case errors.Is(err, memory.ErrNoTracks), errors.Is(err, memory.ErrNoDistricts):
		return nil, status.Error(codes.NotFound, err.Error())
	case errors.Is(err, memory.ErrDistrictNotFound):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.ListAllTracksByDistrictResponse{Tracks: tracks}, nil
}

func (s *Service) ListDriftTracksByDistrict(ctx context.Context, req *pb.ListDriftTracksByDistrictRequest) (*pb.ListDriftTracksByDistrictResponse, error) {
	tracks, err := s.repo.ListDriftTracksByDistrict(ctx, req.GetDistrict())

	switch {
	case errors.Is(err, memory.ErrNoTracks), errors.Is(err, memory.ErrNoDistricts):
		return nil, status.Error(codes.NotFound, err.Error())
	case errors.Is(err, memory.ErrDistrictNotFound):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.ListDriftTracksByDistrictResponse{Tracks: tracks}, nil
}

func (s *Service) GetAlternativesByDistrictAndTrack(ctx context.Context, req *pb.GetAlternativesByDistrictAndTrackRequest) (*pb.GetAlternativesByDistrictAndTrackResponse, error) {
	tracks, err := s.repo.GetAlternativesByDistrictAndTrack(ctx, req.GetDistrict(), req.GetTrack())

	switch {
	case errors.Is(err, memory.ErrNoTracks), errors.Is(err, memory.ErrNoDistricts):
		return nil, status.Error(codes.NotFound, err.Error())
	case errors.Is(err, memory.ErrDistrictNotFound):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.GetAlternativesByDistrictAndTrackResponse{Tracks: tracks}, nil
}

func (s *Service) GetCollisionsByDistrictAndTrack(ctx context.Context, req *pb.GetCollisionsByDistrictAndTrackRequest) (*pb.GetCollisionsByDistrictAndTrackResponse, error) {
	tracks, err := s.repo.GetCollisionsByDistrictAndTrack(ctx, req.GetDistrict(), req.GetTrack())

	switch {
	case errors.Is(err, memory.ErrNoTracks), errors.Is(err, memory.ErrNoDistricts):
		return nil, status.Error(codes.NotFound, err.Error())
	case errors.Is(err, memory.ErrDistrictNotFound):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.GetCollisionsByDistrictAndTrackResponse{Tracks: tracks}, nil
}
