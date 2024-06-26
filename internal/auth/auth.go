package auth

import (
	"context"
	"time"

	pb "real_time_voting_system/internal/auth/proto"
	"real_time_voting_system/internal/storage"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var jwtKey = []byte("my_secret_key")

type AuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
	redisClient *storage.RedisClient
}

func NewAuthServiceServer(redisClient *storage.RedisClient) *AuthServiceServer {
	return &AuthServiceServer{redisClient: redisClient}
}

func (s *AuthServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	err := s.redisClient.SetUser(req.Username, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}
	return &pb.RegisterResponse{Message: "User registered successfully"}, nil
}

func (s *AuthServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	password, err := s.redisClient.GetUser(req.Username)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials: %v", err)
	}
	if password != req.Password {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}
	token, err := generateJWT(req.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error generating token")
	}
	// Store the token in Redis
	err = s.redisClient.SetToken(req.Username, token)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error storing token in Redis")
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
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
