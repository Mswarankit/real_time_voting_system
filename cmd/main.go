package main

import (
	"log"
	"net"
	"net/http"

	"real_time_voting_system/internal/auth"
	"real_time_voting_system/internal/storage"
	"real_time_voting_system/internal/websocket"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize Redis
	redisClient := storage.NewRedisClient()
	if redisClient == nil {
		log.Fatalf("failed to initialize Redis client")
	}

	// Initialize gRPC server
	grpcServer := grpc.NewServer()
	authService := auth.NewAuthServiceServer()
	auth.RegisterAuthServiceServer(grpcServer, authService)
	reflection.Register(grpcServer)

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
	hub := websocket.NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	log.Println("HTTP server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
