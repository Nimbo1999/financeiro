package handler

import (
	"context"
	"log"

	"github.com/nimbo1999/financeiro/users/internal/services"
	userpb "github.com/nimbo1999/financeiro/users/pkg/grpc/users/v1"
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
	// Validate request
	if err := ValidateGetUserByEmailRequest(req); err != nil {
		return nil, err
	}

	user, err := h.userService.GetUserByEmail(req.Email)
	if err != nil {
		// Check if user not found - return found=false instead of error
		if err == services.ErrUserNotFound {
			return &userpb.GetUserByEmailResponse{
				Found: false,
			}, nil
		}
		// Map other errors to appropriate gRPC codes
		return nil, MapServiceErrorToGRPC(err, "GetUserByEmail")
	}

	// Convert domain model to protobuf message
	protoUser := ModelToProto(user)

	return &userpb.GetUserByEmailResponse{
		User:  protoUser,
		Found: true,
	}, nil
}

func (h *UserGRPCHandler) GetUserById(ctx context.Context, req *userpb.GetUserByIdRequest) (*userpb.GetUserByIdResponse, error) {
	// Validate request
	if err := ValidateGetUserByIdRequest(req); err != nil {
		return nil, err
	}

	user, err := h.userService.GetUserByID(req.Id)
	if err != nil {
		// Check if user not found - return found=false instead of error
		if err == services.ErrUserNotFound {
			return &userpb.GetUserByIdResponse{
				Found: false,
			}, nil
		}
		// Map other errors to appropriate gRPC codes
		return nil, MapServiceErrorToGRPC(err, "GetUserById")
	}

	// Convert domain model to protobuf message
	protoUser := ModelToProto(user)

	return &userpb.GetUserByIdResponse{
		User:  protoUser,
		Found: true,
	}, nil
}

func (h *UserGRPCHandler) HealthCheck(ctx context.Context, req *userpb.HealthCheckRequest) (*userpb.HealthCheckResponse, error) {
	// Perform basic health checks

	// Test service availability by attempting a simple operation
	// This indirectly tests database connectivity through the service layer
	_, err := h.userService.ListUsers(&services.PaginationParams{
		Page:     1,
		PageSize: 1,
	})

	if err != nil {
		log.Printf("Health check failed: %v", err)
		return &userpb.HealthCheckResponse{
			Status:  userpb.HealthCheckResponse_NOT_SERVING,
			Message: "Service unavailable - database connectivity issues",
		}, nil
	}

	return &userpb.HealthCheckResponse{
		Status:  userpb.HealthCheckResponse_SERVING,
		Message: "User service is healthy",
	}, nil
}