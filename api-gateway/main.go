package main

import (
	"context"
	"example/gateway/config"
	"example/gateway/proto/greeter"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// func authMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Ekstrakcija Authorization zaglavlja
// 		authHeader := r.Header.Get("Authorization")
// 		log.Println("Authorization Header:", authHeader)
// 		if authHeader == "" {
// 			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
// 			return
// 		}

// 		// Postavljanje vrednosti zaglavlja u kontekst
// 		ctx := context.WithValue(r.Context(), "Authorization", authHeader)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

func main() {
	cfg := config.GetConfig()
	log.Printf("Starting server with config: %+v\n", cfg)

	conn, err := grpc.DialContext(
		context.Background(),
		// cfg.GreeterServiceAddress,
		":8095",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	conn2, err := grpc.DialContext(
		context.Background(),
		// cfg.FollowerServiceAddress,
		":8092", // Adresa novog servisa
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalln("Failed to dial AnotherService server:", err)
	}

	conn4, err := grpc.DialContext(
		context.Background(),
		// cfg.TourServiceAddress,
		":8090",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register Greeter
	client := greeter.NewGreeterServiceClient(conn)
	err = greeter.RegisterGreeterServiceHandlerClient(
		context.Background(),
		gwmux,
		client,
	)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	client2 := greeter.NewFollowerServiceClient(conn2)
	err = greeter.RegisterFollowerServiceHandlerClient(
		context.Background(),
		gwmux,
		client2,
	)
	if err != nil {
		log.Fatalln("Failed to register AnotherService gateway:", err)
	}

	client4 := greeter.NewTourServiceClient(conn4)
	err = greeter.RegisterTourServiceHandlerClient(
		context.Background(),
		gwmux,
		client4,
	)
	if err != nil {
		log.Fatalln("Failed to register AnotherService gateway:", err)
	}

	//handler := authMiddleware(gwmux)

	gwServer := &http.Server{
		Addr:    "localhost:8000",
		Handler: gwmux,
	}

	go func() {
		log.Println("Starting HTTP server at localhost:8000")
		if err := gwServer.ListenAndServe(); err != nil {
			//log.Fatal("server error: ", err)
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, syscall.SIGTERM)

	<-stopCh

	log.Println("Shutting down the server...")
	if err = gwServer.Close(); err != nil {
		log.Fatalln("error while stopping server: ", err)
	}
	log.Println("Server stopped gracefully")
}
