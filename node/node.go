package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	//"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	proto "ITU/grpc"
)

type AuctionServer struct {
	proto.UnimplementedAuctionServiceServer
	mu sync.RWMutex

	// Auction state
	highestBid       int64
	highestBidder    string
	isAuctionOver    bool
	auctionStartTime time.Time
	auctionDuration  time.Duration

	// Node management
	port            string
	isLeader        bool
	leaderPort      string
	replicas        map[string]proto.AuctionServiceClient
	nodeConnections map[string]*grpc.ClientConn
}

func main() {
	log.Println("[Server] Starting auction server...")
	startServer()
}

func startServer() {
	server := &AuctionServer{
		replicas:         make(map[string]proto.AuctionServiceClient),
		nodeConnections:  make(map[string]*grpc.ClientConn),
		highestBid:       0,
		highestBidder:    "",
		isAuctionOver:    false,
		auctionStartTime: time.Now(),
		auctionDuration:  100 * time.Second,
	}

	listener := server.getAvailableListener()
	grpcServer := grpc.NewServer()
	proto.RegisterAuctionServiceServer(grpcServer, server)

	go func() {
		log.Printf("Server starting on port %s", server.port)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	fmt.Printf("Auction started! It will run for %v\n", server.auctionDuration)
	fmt.Println("Type 'connect' to connect to other replicas")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "connect" {
			server.connectToReplicas()
			server.startLeaderElection()
			break
		}
	}

	go server.monitorAuctionTimeout()
	go server.monitorLeader()
	go server.monitorAllNodes() // New: monitor all nodes, not just leader

	select {}
}

func (s *AuctionServer) getAvailableListener() net.Listener {
	ports := []string{":50002", ":50003", ":50004"}

	for _, port := range ports {
		listener, err := net.Listen("tcp", port)
		if err == nil {
			s.port = port
			return listener
		}
	}

	log.Fatal("No available ports found")
	return nil
}

func (s *AuctionServer) connectToReplicas() {
	log.Println("Connecting to replica nodes...")
	allPorts := []string{":50002", ":50003", ":50004"}

	for _, port := range allPorts {
		if port == s.port {
			continue
		}

		conn, err := grpc.NewClient("localhost"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("No server on port %s", port)
			continue
		}

		client := proto.NewAuctionServiceClient(conn)
		s.replicas[port] = client
		s.nodeConnections[port] = conn
		log.Printf("Connected to replica on port %s", port)
	}
}

// FIXED: Proper leader election that excludes failed nodes
func (s *AuctionServer) startLeaderElection() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Starting leader election. Available nodes: %v", s.getAvailablePorts())

	// Start with self as candidate
	candidates := []string{s.port}

	// Add all responsive replicas
	for port, client := range s.replicas {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err := client.Result(ctx, &proto.ResultRequest{ClientID: "election-check"})
		cancel()

		if err == nil {
			candidates = append(candidates, port)
			log.Printf("Node %s is responsive, adding to election", port)
		} else {
			log.Printf("Node %s is not responsive, removing from candidates", port)
			// Remove failed node
			delete(s.replicas, port)
			if conn, exists := s.nodeConnections[port]; exists {
				conn.Close()
				delete(s.nodeConnections, port)
			}
		}
	}

	if len(candidates) == 0 {
		log.Printf("No candidates found, something is wrong")
		return
	}

	// Find highest port among responsive nodes
	newLeader := candidates[0]
	for _, candidate := range candidates {
		if candidate > newLeader {
			newLeader = candidate
		}
	}

	// Update leadership
	s.leaderPort = newLeader
	s.isLeader = (s.port == newLeader)

	if s.isLeader {
		log.Printf("🎉 Node %s elected as new leader!", s.port)
	} else {
		log.Printf("Node %s following leader %s", s.port, s.leaderPort)
	}
}

// NEW: Monitor all nodes, not just the leader
func (s *AuctionServer) monitorAllNodes() {
	ticker := time.NewTicker(3 * time.Second) // Check more frequently
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()

		// Check if current leader is still responsive
		if s.leaderPort != "" && s.leaderPort != s.port {
			leaderClient, exists := s.replicas[s.leaderPort]
			if exists {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				_, err := leaderClient.Result(ctx, &proto.ResultRequest{ClientID: "health-check"})
				cancel()

				if err != nil {
					log.Printf("Leader %s is down! Starting election...", s.leaderPort)
					s.mu.Unlock()
					s.startLeaderElection()
					continue
				}
			} else {
				// Leader not in our list, need election
				log.Printf("Leader %s not in replica list, starting election...", s.leaderPort)
				s.mu.Unlock()
				s.startLeaderElection()
				continue
			}
		}

		s.mu.Unlock()
	}
}

// Get list of available ports for logging
func (s *AuctionServer) getAvailablePorts() []string {
	ports := []string{s.port}
	for port := range s.replicas {
		ports = append(ports, port)
	}
	return ports
}

func (s *AuctionServer) monitorLeader() {
	// This is now handled by monitorAllNodes
	// Keeping it empty to avoid conflicts
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
	}
}

func (s *AuctionServer) Bid(ctx context.Context, req *proto.BidRequest) (*proto.BidResponse, error) {
	s.mu.RLock()

	// If I'm not the leader, forward to leader
	if !s.isLeader && s.leaderPort != "" && s.leaderPort != s.port {
		leaderClient, exists := s.replicas[s.leaderPort]
		s.mu.RUnlock()

		if exists {
			log.Printf("Forwarding bid to leader %s", s.leaderPort)
			resp, err := leaderClient.Bid(ctx, req)
			if err != nil {
				// Leader might be down, start election
				log.Printf("Failed to forward to leader %s: %v", s.leaderPort, err)
				s.startLeaderElection()
				// Retry the bid after election
				return s.Bid(ctx, req)
			}
			return resp, nil
		} else {
			// Leader not available, start election
			log.Printf("Leader %s not available, starting election", s.leaderPort)
			s.startLeaderElection()
			// Retry the bid after election
			return s.Bid(ctx, req)
		}
	} else {
		s.mu.RUnlock()
	}

	// If we get here, we are the leader
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Processing bid locally as leader: %d from client %s", req.Amount, req.ClientID)

	// Check auction state
	elapsed := time.Since(s.auctionStartTime)
	if s.isAuctionOver || elapsed >= s.auctionDuration {
		s.isAuctionOver = true
		return &proto.BidResponse{Ack: "exception: Auction has ended"}, nil
	}

	// Validate bid
	if req.Amount <= s.highestBid {
		return &proto.BidResponse{Ack: "fail: Bid must be higher than current highest bid"}, nil
	}

	// Accept bid
	s.highestBid = req.Amount
	s.highestBidder = req.ClientID

	log.Printf("New highest bid: %d from client %s", req.Amount, req.ClientID)

	// Replicate to other nodes
	s.replicateBidToReplicas(req.ClientID, req.Amount)

	return &proto.BidResponse{Ack: "success: Bid accepted"}, nil
}

func (s *AuctionServer) Result(ctx context.Context, req *proto.ResultRequest) (*proto.ResultResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	elapsed := time.Since(s.auctionStartTime)

	response := &proto.ResultResponse{
		HighestBid:    s.highestBid,
		Winner:        s.highestBidder,
		IsAuctionOver: s.isAuctionOver,
	}

	if s.isAuctionOver || elapsed >= s.auctionDuration {
		response.Status = "finished"
	} else {
		response.Status = "ongoing"
	}

	return response, nil
}

func (s *AuctionServer) Replicate(ctx context.Context, req *proto.ReplicationRequest) (*proto.ReplicationResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.Operation == "bid" {
		if req.Amount > s.highestBid {
			s.highestBid = req.Amount
			s.highestBidder = req.ClientID
			log.Printf("Replicated bid: %d from client %s", req.Amount, req.ClientID)
		}
	} else if req.Operation == "state" {
		s.highestBid = req.HighestBid
		s.highestBidder = req.Winner
		s.isAuctionOver = req.IsAuctionOver
	}

	return &proto.ReplicationResponse{Success: true}, nil
}

func (s *AuctionServer) replicateBidToReplicas(clientID string, amount int64) {
	replicationReq := &proto.ReplicationRequest{
		Operation: "bid",
		ClientID:  clientID,
		Amount:    amount,
	}

	successfulReplications := 0
	for port, client := range s.replicas {
		go func(p string, c proto.AuctionServiceClient) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			_, err := c.Replicate(ctx, replicationReq)
			if err != nil {
				log.Printf("Could not replicate to %s - marking as failed", p)
				// Mark this node as failed
				s.mu.Lock()
				delete(s.replicas, p)
				if conn, exists := s.nodeConnections[p]; exists {
					conn.Close()
					delete(s.nodeConnections, p)
				}
				s.mu.Unlock()
			} else {
				successfulReplications++
			}
		}(port, client)
	}

	log.Printf("Replication: %d/%d successful", successfulReplications, len(s.replicas))
}

func (s *AuctionServer) replicateAuctionEnd() {
	replicationReq := &proto.ReplicationRequest{
		Operation:     "state",
		HighestBid:    s.highestBid,
		Winner:        s.highestBidder,
		IsAuctionOver: true,
	}

	for port, client := range s.replicas {
		go func(p string, c proto.AuctionServiceClient) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			c.Replicate(ctx, replicationReq)
		}(port, client)
	}
}

func (s *AuctionServer) monitorAuctionTimeout() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		elapsed := time.Since(s.auctionStartTime)
		if !s.isAuctionOver && elapsed >= s.auctionDuration {
			s.isAuctionOver = true
			log.Printf("Auction ended! Winner: %s with bid: %d", s.highestBidder, s.highestBid)
			if s.isLeader {
				s.replicateAuctionEnd()
			}
		}
		s.mu.Unlock()
	}
}
