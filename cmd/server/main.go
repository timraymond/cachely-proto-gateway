package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	cachelyv1 "github.com/timraymond/cachely/cachelyv1/cachely/v1"
)

type server struct {
	data sync.Map
}

func (s *server) Get(ctx context.Context, req *cachelyv1.GetRequest) (*cachelyv1.GetResponse, error) {
	key := req.GetKey()
	log.Printf("looking up key %q\n", key)
	if v, ok := s.data.Load(key); ok {
		log.Printf("found key %q\n", key)
		return &cachelyv1.GetResponse{
			Key:   key,
			Value: v.([]byte),
		}, status.New(codes.OK, "").Err()
	}
	log.Printf("key not found %q\n", key)
	return nil, status.Errorf(codes.NotFound, "could not find key %s", key)
}

// Delete will remove the cached value located at key from the cache. If there
// is no value at the provided key, an error will be produced.
func (s *server) Delete(ctx context.Context, req *cachelyv1.DeleteRequest) (*cachelyv1.DeleteResponse, error) {
	key := req.GetKey()
	log.Printf("Storing key: %q\n", key)

	if _, ok := s.data.Load(key); ok {
		s.data.Delete(key)

		return &cachelyv1.DeleteResponse{
			Key: key,
		}, nil
	}

	return nil, status.Errorf(codes.NotFound, "could not find key %s", key)
}

// Put stores the provided value at the key specified. If there is an existing
// entry, it will return an error. Delete should be called on that entry first.
func (s *server) Put(ctx context.Context, req *cachelyv1.PutRequest) (*cachelyv1.PutResponse, error) {
	key := req.GetKey()
	val := req.GetValue()

	log.Printf("Writing value at key: %q\n", key)

	if _, ok := s.data.Load(key); !ok {
		s.data.Store(key, val)

		return &cachelyv1.PutResponse{
			Key: key,
		}, nil
	}

	return nil, status.Errorf(codes.AlreadyExists, "existing cached item located at %s", key)
}

func main() {
	sock, err := net.Listen("tcp", ":5051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create a new grpc server
	s := grpc.NewServer()

	// #TODO: create our new server. Make sure to provide it a sync.Map{}
	srv := &server{
		data: sync.Map{},
	}

	// #TODO: Register the new server by calling `cachely.RegisterCacheServer`
	cachelyv1.RegisterCacheAPIServer(s, srv)

	// enable reflection
	reflection.Register(s)

	// Let the world know we are starting and where we are listening
	log.Printf("starting gRPC service on %s\n", sock.Addr())

	// setup the gateway
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = cachelyv1.RegisterCacheAPIHandlerFromEndpoint(context.TODO(), mux, sock.Addr().String(), opts)
	if err != nil {
		log.Println("err starting grpc gateway: err:", err)
		return
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	wg := sync.WaitGroup{}

	go func() {
		wg.Add(1)
		defer wg.Done()
		// start listening and responding
		if err := s.Serve(sock); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()

		err := http.ListenAndServe(":8080", mux)
		if err != http.ErrServerClosed {
			log.Fatalf("failed to serve: %v\n", err)
		}
	}()

	<-sig
	log.Println("Shutdown signal received. Starting graceful shutdown")
}
