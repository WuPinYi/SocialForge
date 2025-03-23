package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/WuPinYi/SocialForge/internal/auth"
	"github.com/WuPinYi/SocialForge/internal/ent"
	"github.com/WuPinYi/SocialForge/internal/server"
	"github.com/WuPinYi/SocialForge/internal/worker"
	ocsv1 "github.com/WuPinYi/SocialForge/proto/ocs/v1"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize database connection
	client, err := ent.Open("postgres", "host=localhost port=5432 user=postgres dbname=socialforge password=postgres sslmode=disable")
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	// Run the auto migration tool
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	// Create Auth0 middleware
	auth0Config := auth.Auth0Config{
		Domain: os.Getenv("AUTH0_DOMAIN"),
	}
	auth0Middleware, err := auth.NewAuth0Middleware(auth0Config)
	if err != nil {
		log.Fatalf("failed creating Auth0 middleware: %v", err)
	}

	// Create gRPC server
	s := grpc.NewServer(
		grpc.UnaryInterceptor(auth0Middleware.UnaryInterceptor),
	)
	ocsv1.RegisterOpinionControlServiceServer(s, server.NewServer(client))

	// Register reflection service for development
	reflection.Register(s)

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the post worker
	postWorker := worker.NewPostWorker(client)
	go postWorker.Start(ctx)

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down gRPC server...")
		cancel()
		s.GracefulStop()
	}()

	// Start listening
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
