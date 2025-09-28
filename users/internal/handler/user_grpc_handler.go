package handler

import (
	"context"

	"github.com/nimbo1999/financeiro/users/internal/services"
	userpb "github.com/nimbo1999/financeiro/users/pkg/grpc/users/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserGRPCHandler struct {
	userpb.UnimplementedUserServiceServer
	userService services.UserService
}

func NewUserGRPCHandler(userService services.UserService) *UserGRPCHandler {
	return &UserGRPCHandler{
		userService: userService,
	}
}

func (h *UserGRPCHandler) GetUserByEmail(ctx context.Context, req *userpb.GetUserByEmailRequest) (*userpb.GetUserByEmailResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	user, err := h.userService.GetUserByEmail(req.Email)
	if err != nil {
		// Check if user not found
		if err == services.ErrUserNotFound {
			return &userpb.GetUserByEmailResponse{
				Found: false,
			}, nil
		}
		return nil, status.Error(codes.Internal, "failed to get user by email")
	}

	protoUser := &userpb.User{
		Id:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}

	return &userpb.GetUserByEmailResponse{
		User:  protoUser,
		Found: true,
	}, nil
}

func (h *UserGRPCHandler) GetUserById(ctx context.Context, req *userpb.GetUserByIdRequest) (*userpb.GetUserByIdResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	user, err := h.userService.GetUserByID(req.Id)
	if err != nil {
		// Check if user not found
		if err == services.ErrUserNotFound {
			return &userpb.GetUserByIdResponse{
				Found: false,
			}, nil
		}
		return nil, status.Error(codes.Internal, "failed to get user by ID")
	}

	protoUser := &userpb.User{
		Id:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}

	return &userpb.GetUserByIdResponse{
		User:  protoUser,
		Found: true,
	}, nil
}

func (h *UserGRPCHandler) HealthCheck(ctx context.Context, req *userpb.HealthCheckRequest) (*userpb.HealthCheckResponse, error) {
	// Simple health check - could be expanded to check database connectivity, etc.
	return &userpb.HealthCheckResponse{
		Status:  userpb.HealthCheckResponse_SERVING,
		Message: "User service is healthy",
	}, nil
}