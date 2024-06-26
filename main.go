package main

import (
	"log"
	"net"
	"net/http"

	"real_time_voting_system/internal/auth"
	pb "real_time_voting_system/internal/auth/proto"
	"real_time_voting_system/internal/storage"
	"real_time_voting_system/internal/websocket"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize Redis
	redisClient := storage.NewRedisClient()

	// Initialize gRPC server
	grpcServer := grpc.NewServer()
	authService := auth.NewAuthServiceServer(redisClient)
	pb.RegisterAuthServiceServer(grpcServer, authService)
	reflection.Register(grpcServer)

	// Start the gRPC server in a separate goroutine
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		log.Println("gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Initialize WebSocket hub
	hub := websocket.NewHub(redisClient)
	go hub.Run()

	// Handle WebSocket connections
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	// Start the HTTP server
	log.Println("HTTP server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
