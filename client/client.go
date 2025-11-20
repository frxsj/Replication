package main

import (
	proto "ITU/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type gangClient struct {
	client proto.GangnameStyleClient
	id string
	
}

func main() {
	gangClient{
		client : connectClient()
	}  // How do we randomize the ID?

	reader := bufio.NewReader(os.Stdin)
	gangClient.client.initiateAuctionTalk(reader)

}

func (gang *GangnameStyleClient) initiateAuctionTalk(reader *bufio.Reader) {
	fmt.Println("Welcome to the auction!")
	fmt.Println("Today's item is: one (1) Arla Cultura, strawberry flavor")
	fmt.Println("Type /bid and the amount you would like to bid on the item!")
	fmt.Println("Type /result to get the current status of the auction.")

	for {
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)

		if message == "/result" {
			plsres := &proto.PlsResult{
				ClientID: gang.id,
			}
			gang.Result(context.Background(),plsres)
		}
		
		if strings.HasPrefix(message, "/bid") {
			var bid = strings.Split(message, " ")
			if(len(bid)> 2) {
				fmt.Println("[Failure] Too many words.")
			} else if(len(bid) <= 1) {
				fmt.Println("[Failure] Don't forget the amount!")
			} else {
				amount, err := strconv.ParseInt(bid[1], 10, 64) // String, Base 10, int64
				if(err != nil) {
					fmt.Println("You cant do that my drilla.")
					return
				}
				plsBid := &proto.Request{
					ClientID : gang.id,
					Amount : amount,
				}
				response, err := gang.client.Bid(context.Background(), plsBid)
				if err != nil {
					fmt.Println("[Client] Failed to send!")
				}
				return response
			}

		}

	}

}

func connectClient()(proto.NewGangnameStyleClient) {
	conn, err := grpc.NewClient("localhost:5005", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working")
	}

	client := proto.NewGangnameStyleClient(conn)
	return client
}

func (c *gangClient) Result(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	return &proto.Response{}, nil
}