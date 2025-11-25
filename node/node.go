package main

import (
	proto "ITU/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type gangnamStyle struct {
	proto.UnimplementedGangnamStyleServer
	mu sync.RWMutex

	// Auction state
	highestBid       int64
	highestBidder    string
	isAuctionOver    bool
	auctionStartTime time.Time
	auctionDuration  time.Duration

	// Node management
	port            string
	amIsILeader     bool
	leaderPort      string
	replicas        map[string]proto.GangnamStyleClient
	biddersMap      map[string]*grpc.ClientConn
	nodeConnections map[string]*grpc.ClientConn
}

func main() {
	log.Println("[Server] Starting server...")
	start_server()
}

func start_server() {
	gangy := &gangnamStyle{
		replicas:         make(map[string]proto.GangnamStyleClient),
		biddersMap:       make(map[string]*grpc.ClientConn),
		nodeConnections:  make(map[string]*grpc.ClientConn),
		highestBid:       0,
		highestBidder:    "",
		isAuctionOver:    false,
		auctionStartTime: time.Now(),
		auctionDuration:  100 * time.Second,
	}

	node := grpc.NewServer()
	proto.RegisterGangnamStyleServer(node, gangy)
	listener := gangy.get_available_listener()

	go func() {
		if err := node.Serve(listener); err != nil {
			log.Fatalf("womp womp): %v", err)
		}
	}()

	fmt.Println("Once all replicas have been started type 'arla cultura 4eva<3<3'")
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		if scanner.Text() == "arla cultura 4eva<3<3" {
			gangy.connect_nodes()
			gangy.selectGangLeader()
			break
		}
	}

	go gangy.monitorAuctionTimeout()
	go gangy.monitorLeader()

	// Keep the server running
	select {}
}

func (gangy *gangnamStyle) get_available_listener() net.Listener {
	ports := []string{":50002", ":50003", ":50004"}

	for _, port := range ports {
		listener, err := net.Listen("tcp", port)
		if err == nil {
			fmt.Println("got port", port)
			gangy.port = port
			return listener
		}
	}

	log.Fatalf("left a lot of crumbs")
	return nil
}

func (gangy *gangnamStyle) connect_nodes() {
	fmt.Println("connecting the reaplicas...")
	ports := []string{":50002", ":50003", ":50004"}

	for _, port := range ports {
		if port == gangy.port {
			continue
		}

		conn, err := grpc.NewClient(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Failed to slay to %s: %v", port, err)
			continue
		}

		client := proto.NewGangnamStyleClient(conn)
		gangy.replicas[port] = client
		gangy.nodeConnections[port] = conn
		fmt.Printf("Connected to replica on port %s\n", port)
	}
}

func (gangy *gangnamStyle) selectGangLeader() {
	fmt.Println("banke banke på")

	gangy.mu.Lock()
	defer gangy.mu.Unlock()

	// Check which nodes are actually responsive
	responsiveNodes := []string{gangy.port} // Always include self

	for port, client := range gangy.replicas {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err := client.Result(ctx, &proto.PlsResult{ClientID: "election-check"})
		cancel()

		if err == nil {
			responsiveNodes = append(responsiveNodes, port)
			fmt.Printf("Node %s is responsive\n", port)
		} else {
			fmt.Printf("Node %s is NOT responsive - removing from election\n", port)
			// Remove failed node
			delete(gangy.replicas, port)
			if conn, exists := gangy.nodeConnections[port]; exists {
				conn.Close()
				delete(gangy.nodeConnections, port)
			}
		}
	}

	// Find highest port among responsive nodes
	newLeader := responsiveNodes[0]
	for _, node := range responsiveNodes {
		if node > newLeader {
			newLeader = node
		}
	}

	// Update leadership
	gangy.leaderPort = newLeader
	gangy.amIsILeader = (gangy.port == newLeader)

	if gangy.amIsILeader {
		fmt.Printf("🎉 The new alpha has been chosen🐺 on port %s\n", gangy.port)
	} else {
		fmt.Printf("Node %s following leader %s\n", gangy.port, gangy.leaderPort)
	}
}

func (gangy *gangnamStyle) monitorLeader() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !gangy.amIsILeader && gangy.leaderPort != "" {
			leaderClient, exists := gangy.replicas[gangy.leaderPort]
			if exists {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				_, err := leaderClient.Result(ctx, &proto.PlsResult{ClientID: "health-check"})
				cancel()

				if err != nil {
					fmt.Printf("Leader %s appears to be down! Starting election...\n", gangy.leaderPort)
					gangy.selectGangLeader()
				}
			} else {
				// Leader not in our replica list anymore
				fmt.Printf("Leader %s not in replica list! Starting election...\n", gangy.leaderPort)
				gangy.selectGangLeader()
			}
		}
	}
}

func (gangy *gangnamStyle) Bid(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	gangy.mu.Lock()
	defer gangy.mu.Unlock()

	fmt.Printf("Received bid: %d from client %s\n", req.Amount, req.ClientID)

	// Check if auction is over
	elapsed := time.Since(gangy.auctionStartTime)
	if gangy.isAuctionOver || elapsed >= gangy.auctionDuration {
		gangy.isAuctionOver = true
		return &proto.Response{Ack: "exception: Auction has ended"}, nil
	}

	// Validate bid
	if req.Amount <= gangy.highestBid {
		return &proto.Response{Ack: "fail: Bid must be higher than current highest bid"}, nil
	}

	// Accept bid
	gangy.highestBid = req.Amount
	gangy.highestBidder = req.ClientID

	fmt.Printf("New highest bid: %d from client %s\n", req.Amount, req.ClientID)

	// Replicate to other nodes if leader
	if gangy.amIsILeader {
		gangy.forwardToAlpha(req)
	}

	return &proto.Response{Ack: "Your bet was registered, my drilla😎"}, nil
}

func (gangy *gangnamStyle) Result(ctx context.Context, req *proto.PlsResult) (*proto.Outcome, error) {
	gangy.mu.RLock()
	defer gangy.mu.RUnlock()

	outcome := &proto.Outcome{
		Winner:       gangy.highestBidder,
		WinnerAmount: gangy.highestBid,
	}

	return outcome, nil
}

func (gangy *gangnamStyle) Bully(ctx context.Context, b_req *proto.BullyRequest) (*proto.BullyResponse, error) {
	fmt.Println("The bullying has begun!")

	// Simple replication - update state from leader
	if b_req.Senderport != gangy.port {
		fmt.Printf("Received replication from %s\n", b_req.Senderport)
	}

	return &proto.BullyResponse{AlphaPort: gangy.leaderPort}, nil
}

func (gangy *gangnamStyle) forwardToAlpha(req *proto.Request) {
	fmt.Println("Forwarding to alpha nodes...")

	successCount := 0
	for port, client := range gangy.replicas {
		go func(p string, c proto.GangnamStyleClient) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			_, err := c.Bully(ctx, &proto.BullyRequest{Senderport: gangy.port})
			if err != nil {
				fmt.Printf("Failed to forward to %s: %v\n", p, err)
				// Mark this node as failed
				gangy.mu.Lock()
				delete(gangy.replicas, p)
				if conn, exists := gangy.nodeConnections[p]; exists {
					conn.Close()
					delete(gangy.nodeConnections, p)
				}
				gangy.mu.Unlock()
			} else {
				successCount++
				fmt.Printf("Successfully forwarded to %s\n", p)
			}
		}(port, client)
	}

	fmt.Printf("Replication: %d/%d successful\n", successCount, len(gangy.replicas))
}

func (gangy *gangnamStyle) monitorAuctionTimeout() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		gangy.mu.Lock()
		elapsed := time.Since(gangy.auctionStartTime)
		if !gangy.isAuctionOver && elapsed >= gangy.auctionDuration {
			gangy.isAuctionOver = true
			fmt.Printf("Auction ended! Winner: %s with bid: %d\n", gangy.highestBidder, gangy.highestBid)
		}
		gangy.mu.Unlock()
	}
}
