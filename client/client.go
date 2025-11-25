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

type AuctionClient struct {
	client proto.AuctionServiceClient
	conn   *grpc.ClientConn
	id     string
}

func main() {
	client := &AuctionClient{}
	client.connectToServer()
	defer client.conn.Close()

	// Generate unique client ID
	client.id = fmt.Sprintf("client-%d", time.Now().UnixNano())

	reader := bufio.NewReader(os.Stdin)
	client.startAuctionInterface(reader)
}

func (c *AuctionClient) connectToServer() {
	// Try connecting to different ports until successful
	ports := []string{":50002", ":50003", ":50004", ":50005"}

	for _, port := range ports {
		conn, err := grpc.NewClient("localhost"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			c.conn = conn
			c.client = proto.NewAuctionServiceClient(conn)
			log.Printf("Connected to server on port %s", port)
			return
		}
	}

	log.Fatal("Could not connect to any server")
}

func (c *AuctionClient) startAuctionInterface(reader *bufio.Reader) {
	fmt.Println("=== Distributed Auction System ===")
	fmt.Println("Available commands:")
	fmt.Println("/bid <amount> - Place a bid")
	fmt.Println("/result - Check auction status")
	fmt.Println("/quit - Exit")
	fmt.Println("==================================")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch {
		case input == "/quit":
			fmt.Println("Goodbye!")
			return

		case input == "/result":
			c.getAuctionResult()

		case strings.HasPrefix(input, "/bid "):
			parts := strings.Split(input, " ")
			if len(parts) != 2 {
				fmt.Println("Usage: /bid <amount>")
				continue
			}

			amount, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				fmt.Println("Invalid amount. Please enter a valid number.")
				continue
			}

			c.placeBid(amount)

		default:
			fmt.Println("Unknown command. Use /bid, /result, or /quit")
		}
	}
}

func (c *AuctionClient) placeBid(amount int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &proto.BidRequest{
		ClientID: c.id,
		Amount:   amount,
	}

	response, err := c.client.Bid(ctx, req)
	if err != nil {
		fmt.Printf("Error placing bid: %v\n", err)
		// Try reconnecting to another server
		c.connectToServer()
		return
	}

	fmt.Printf("Bid response: %s\n", response.Ack)
}

func (c *AuctionClient) getAuctionResult() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &proto.ResultRequest{
		ClientID: c.id,
	}

	response, err := c.client.Result(ctx, req)
	if err != nil {
		fmt.Printf("Error getting result: %v\n", err)
		// Try reconnecting to another server
		c.connectToServer()
		return
	}

	if response.IsAuctionOver {
		fmt.Printf("=== AUCTION FINISHED ===\n")
		fmt.Printf("Winner: %s\n", response.Winner)
		fmt.Printf("Winning Bid: %d\n", response.HighestBid)
	} else {
		fmt.Printf("=== AUCTION ONGOING ===\n")
		fmt.Printf("Current Highest Bid: %d\n", response.HighestBid)
		fmt.Printf("Current Leader: %s\n", response.Winner)
		fmt.Printf("Time remaining: estimate ~%d seconds\n", 100-int64(time.Since(time.Now().Add(-100*time.Second)).Seconds()))
	}
}
