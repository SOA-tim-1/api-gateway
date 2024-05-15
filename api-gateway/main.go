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

	conn, err := grpc.DialContext(
		context.Background(),
		cfg.GreeterServiceAddress,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	conn2, err := grpc.DialContext(
		context.Background(),
		cfg.FollowerServiceAddress, // Adresa novog servisa
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalln("Failed to dial AnotherService server:", err)
	}

	conn3, err := grpc.DialContext(
		context.Background(),
		cfg.StakeholderServiceAddress, // Adresa novog servisa
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalln("Failed to dial AnotherService server:", err)
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

	client3 := greeter.NewAuthorizeClient(conn3)
	err = greeter.RegisterAuthorizeHandlerClient(
		context.Background(),
		gwmux,
		client3,
	)
	if err != nil {
		log.Fatalln("Failed to register AnotherService gateway:", err)
	}

	client4 := greeter.NewUserServiceClient(conn3)
	err = greeter.RegisterUserServiceHandlerClient(
		context.Background(),
		gwmux,
		client4,
	)
	if err != nil {
		log.Fatalln("Failed to register AnotherService gateway:", err)
	}

	client5 := greeter.NewPersonServiceClient(conn3)
	err = greeter.RegisterPersonServiceHandlerClient(
		context.Background(),
		gwmux,
		client5,
	)
	if err != nil {
		log.Fatalln("Failed to register AnotherService gateway:", err)
	}
	//handler := authMiddleware(gwmux)

	gwServer := &http.Server{
		Addr:    cfg.Address,
		Handler: gwmux,
	}

	go func() {
		if err := gwServer.ListenAndServe(); err != nil {
			log.Fatal("server error: ", err)
		}
	}()

	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, syscall.SIGTERM)

	<-stopCh

	if err = gwServer.Close(); err != nil {
		log.Fatalln("error while stopping server: ", err)
	}
}
