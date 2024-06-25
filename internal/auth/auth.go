
package auth

import (
	"context"
	"time"
	
	pb "real_time_voting_system/internal/auth/proto"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var jwtKey = []byte("my_secret_key")

type AuthServiceServer struct {
	// UnimplementedAuthServiceServer
	users map[string]string
}

func NewAuthServiceServer() *AuthServiceServer {
	return &AuthServiceServer{users: make(map[string]string)}
}

func (s *AuthServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if _, exists := s.users[req.Username]; exists {
		return nil, status.Errorf(codes.AlreadyExists, "User already exists")
	}
	s.users[req.Username] = req.Password
	return &pb.RegisterResponse{Message: "User registered successfully"}, nil
}

func (s *AuthServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if password, exists := s.users[req.Username]; !exists || password != req.Password {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid credentials")
	}
	token, err := generateJWT(req.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error generating token")
	}
	return &pb.LoginResponse{Token: token}, nil
}

func (s *AuthServiceServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	// Implement token invalidation logic if needed
	return &pb.LogoutResponse{Message: "User logged out successfully"}, nil
}

func generateJWT(username string) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
