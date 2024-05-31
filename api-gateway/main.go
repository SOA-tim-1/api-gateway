package main

import (
	"context"
	"crypto/tls"
	"example/gateway/config"
	"example/gateway/proto/greeter"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "[api-gateway] ", log.LstdFlags),
	}
}

// LoggingMiddleware logs each incoming HTTP request.
func (l *Logger) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		l.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		l.Printf("Processed request: %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// logGRPC logs the details of each gRPC call.
func (l *Logger) logGRPC(ctx context.Context, clientName, methodName string, call func(ctx context.Context) error) error {
	start := time.Now()
	l.Printf("Calling gRPC method: %s.%s", clientName, methodName)
	err := call(ctx)
	if err != nil {
		l.Printf("gRPC method %s.%s failed: %v", clientName, methodName, err)
	} else {
		l.Printf("gRPC method %s.%s succeeded in %v", clientName, methodName, time.Since(start))
	}
	return err
}

func combinedCORSHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			// Respond to OPTIONS requests with the appropriate CORS headers
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continue handling the request
		next.ServeHTTP(w, r)
	})
}

func main() {
	logger := NewLogger()
	logger.Println("Fetching configuration...")

	cfg := config.GetConfig()
	logger.Printf("Configuration loaded: %+v\n", cfg)

	logger.Println("Creating gRPC connection to the Greeting service...")
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

	logger.Println("Creating gRPC connection to the Follower service...")
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

	logger.Println("Creating gRPC connection to the Tour service...")
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

	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	creds := credentials.NewTLS(config)

	logger.Println("Creating gRPC connection to the Stakeholders service...")
	conn5, err := grpc.DialContext(
		context.Background(),
		":44332",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(creds),
	)

	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	logger.Println("Creating gRPC connection to the Blog service...")
	conn6, err := grpc.DialContext(
		context.Background(),
		":44333",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(creds),
	)

	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	logger.Println("Creating ServeMux for gRPC Gateway...")
	gwmux := runtime.NewServeMux()

	client := greeter.NewGreeterServiceClient(conn)

	logger.Println("Registering Greeting service handler...")
	err = greeter.RegisterGreeterServiceHandlerClient(
		context.Background(),
		gwmux,
		client,
	)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	client2 := greeter.NewFollowerServiceClient(conn2)

	logger.Println("Registering Follower service handler...")
	err = greeter.RegisterFollowerServiceHandlerClient(
		context.Background(),
		gwmux,
		client2,
	)
	if err != nil {
		log.Fatalln("Failed to register AnotherService gateway:", err)
	}

	client4 := greeter.NewTourServiceClient(conn4)

	logger.Println("Registering Tour service handler...")
	err = greeter.RegisterTourServiceHandlerClient(
		context.Background(),
		gwmux,
		client4,
	)
	if err != nil {
		log.Fatalln("Failed to register AnotherService gateway:", err)
	}

	client5 := greeter.NewCheckpointServiceClient(conn4)
	logger.Println("Registering Checkpoint service handler...")
	err = greeter.RegisterCheckpointServiceHandlerClient(
		context.Background(),
		gwmux,
		client5,
	)
	if err != nil {
		log.Fatalln("Failed to register AnotherService gateway:", err)
	}

	client6 := greeter.NewAuthorizeClient(conn5)
	logger.Println("Registering Authorize service handler...")
	err = greeter.RegisterAuthorizeHandlerClient(
		context.Background(),
		gwmux,
		client6,
	)
	if err != nil {
		log.Fatalln("Failed to register Authorize Service gateway:", err)
	}

	client7 := greeter.NewUserServiceClient(conn5)
	logger.Println("Registering User service handler...")
	err = greeter.RegisterUserServiceHandlerClient(
		context.Background(),
		gwmux,
		client7,
	)
	if err != nil {
		log.Fatalln("Failed to register User Service gateway:", err)
	}

	client8 := greeter.NewPersonServiceClient(conn5)
	logger.Println("Registering Person service handler...")
	err = greeter.RegisterPersonServiceHandlerClient(
		context.Background(),
		gwmux,
		client8,
	)
	if err != nil {
		log.Fatalln("Failed to register Person Service gateway:", err)
	}

	client9 := greeter.NewEquipmentServiceClient(conn4)
	logger.Println("Registering equipment service handler...")
	err = greeter.RegisterEquipmentServiceHandlerClient(
		context.Background(),
		gwmux,
		client9,
	)
	if err != nil {
		log.Fatalln("Failed to register Person Service gateway:", err)
	}

	client10 := greeter.NewCouponServiceClient(conn4)
	logger.Println("Registering coupon service handler...")
	err = greeter.RegisterCouponServiceHandlerClient(
		context.Background(),
		gwmux,
		client10,
	)
	if err != nil {
		log.Fatalln("Failed to register Person Service gateway:", err)
	}

	client11 := greeter.NewBlogServiceClient(conn6)
	logger.Println("Registering blog service handler...")
	err = greeter.RegisterBlogServiceHandlerClient(
		context.Background(),
		gwmux,
		client11,
	)
	if err != nil {
		log.Fatalln("Failed to register Blog Service gateway:", err)
	}

	//handler := authMiddleware(gwmux)

	logger.Println("Adding CORS middleware...")
	corsHandler := combinedCORSHandler(gwmux)
	loggingHandler := logger.LoggingMiddleware(corsHandler) // Add logging middleware
	http.Handle("/", loggingHandler)

	gwServer := &http.Server{
		Addr:    "localhost:8000",
		Handler: loggingHandler,
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
