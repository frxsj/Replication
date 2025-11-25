package main

import (
	proto "ITU/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type gangClient struct {
	client proto.GangnamStyleClient
	id     string
}

func main() {
	gang := gangClient{
		client: connectClient(),
		id:     "client-" + strconv.Itoa(int(time.Now().Unix())),
	}

	fmt.Printf("Client ID: %s\n", gang.id)

	reader := bufio.NewReader(os.Stdin)
	gang.initiateAuctionTalk(reader)
}

func (gang *gangClient) initiateAuctionTalk(reader *bufio.Reader) {
	fmt.Println("Welcome to the auction!😎")
	fmt.Println("Today's item is: one (1) Arla Cultura, strawberry flavor🍶🍓😎")
	fmt.Println("Type /bid and the amount you would like to bid on the item!")
	fmt.Println("Type /result to get the current status of the auction.")

	for {
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)

		if message == "/result" {
			plsres := &proto.PlsResult{
				ClientID: gang.id,
			}
			response, err := gang.client.Result(context.Background(), plsres)
			if err != nil {
				fmt.Println("rip😎")
			} else {
				fmt.Printf("Auction result - Winner: %s, Amount: %d\n", response.Winner, response.WinnerAmount)
			}
		}

		if strings.HasPrefix(message, "/bid") {
			var bid = strings.Split(message, " ")
			if len(bid) > 2 {
				fmt.Println("[Failure] Too many words...")
			} else if len(bid) <= 1 {
				fmt.Println("[Failure] Don't forget the amount!")
			} else {
				amount, err := strconv.ParseInt(bid[1], 10, 64)
				if err != nil {
					fmt.Println("[Client] You cant do that my drilla 😎")
				} else {
					plsBid := &proto.Request{
						ClientID: gang.id,
						Amount:   amount,
					}
					response, err := gang.client.Bid(context.Background(), plsBid)
					if err != nil {
						fmt.Println("[Client] Failed to send!")
					} else {
						fmt.Println(response.Ack)
					}
				}
			}
		}

	}
}

func connectClient() proto.GangnamStyleClient {
	// Try different ports
	ports := []string{":50002", ":50003", ":50004"}

	for _, port := range ports {
		conn, err := grpc.NewClient("localhost"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			fmt.Printf("Connected to server on port %s\n", port)
			client := proto.NewGangnamStyleClient(conn)
			return client
		}
	}

	log.Fatalf("Not working - could not connect to any server")
	return nil
}
