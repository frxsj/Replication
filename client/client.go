package main

import (
	proto "ITU/grpc"
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type gangClient struct {
	client proto.GangnameStyleClient
}

func main() {
	connectClient()

}

func connectClient() {
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working")
	}

	client := proto.NewGangnameStyleClient(conn)

}

func (c *gangnameStyleClient) SendRequest(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {

	return &proto.Response{}, nil
}
