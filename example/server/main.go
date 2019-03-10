package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/horechek/ambassador"
	"github.com/horechek/ambassador/example/pb"
)

var (
	host = flag.String("host", "127.0.0.1", "host")
	port = flag.Int("port", 2379, "port")
)

func main() {
	flag.Parse()

	addr := fmt.Sprintf(*host+":%d", *port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("failed to listen", err)
	}
	defer lis.Close()

	s := grpc.NewServer()
	defer s.GracefulStop()

	pb.RegisterHelloServiceServer(s, &hello{addr: addr})

	service, err := ambassador.NewService("hello", *host, *port, "127.0.0.1:8500", 10)
	if err != nil {
		log.Fatal("failed to listen", err)
	}

	go func() {
		if err := service.Register(); err != nil {
			log.Println("error on deregister", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch

		service.Deregister()

		if i, ok := s.(syscall.Signal); ok {
			os.Exit(int(i))
		} else {
			os.Exit(0)
		}

	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

type hello struct {
	addr string
}

func (h *hello) Echo(ctx context.Context, req *pb.Payload) (*pb.Payload, error) {
	req.Data = req.Data + ", from:" + h.addr
	return req, nil
}
