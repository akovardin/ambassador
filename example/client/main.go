package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"

	"github.com/horechek/ambassador"
	"github.com/horechek/ambassador/example/pb"
)

func main() {
	r := ambassador.NewBuilder("127.0.0.1:8500")

	resolver.Register(r)

	conn, err := grpc.Dial(r.Scheme()+"://author/hello", grpc.WithBalancerName(roundrobin.Name), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	client := pb.NewHelloServiceClient(conn)

	for {
		resp, err := client.Echo(context.Background(), &pb.Payload{Data: "hello"}, grpc.WaitForReady(true))
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(resp)
		}

		<-time.After(time.Second)
	}
}
