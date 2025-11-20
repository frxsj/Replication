package main

import (
	proto "ITU/grpc"
	"context"
	"net"
	"log"
	"fmt"
	"strconv"

	"google.golang.org/grpc"
)

//Leader-based implementation, nodes are replicas
//Like the alpha dawgs

type gangnamStyle struct {
	proto.UnimplementedGangnamStyleServer
	portPack []string
	port string
	amIsILeader bool           //Which replica is the leader
	highestBid int64
	winner string
}

func main() {
	server := &gangnamStyle{}

	var portPack []string{
	"localhost:5005",
    "localhost:5006", 
    "localhost:5007",



	for _, wolf := range wolfPack{
		fmt.Println(wolf.amIsILeader)
	}
	log.Println("[Server] Starting server...")
	server.start_server(port)
	
}

func (gangy *gangnamStyle) forwardToAlpha(req *proto.Request) {
	
}


func selectGangLeader(wolfPack []gangnamStyle) { // Handles leader failure.
	var isAlphaExist bool // Does leader exist
	for i := 0; i < len(wolfPack); i++ {
		if wolfPack[i].amIsILeader { // Goes through the whole list of nodes
			isAlphaExist = true // If the leader is still working the for loop breaks.
			break
		}
	}
	if !isAlphaExist {
		wolfPack[0].amIsILeader = true // Sets first in list as new leader
	}
}

func (gangy *gangnamStyle) start_server(port string){
	node := grpc.NewServer()
	listener, err := net.Listen("tcp", port)
	if(err != nil){
		log.Fatalf("Did not work")
	}

	proto.RegisterGangnamStyleServer(node, gangy)

	log.Println("Server startup was successful at port" + port)
	
	err = node.Serve(listener)
}

func (gangy *gangnamStyle) bid (ctx context.Context, req *proto.Request){
	fmt.Println("you made a bid my drilla")
}

func(gangy gangnamStyle) result (ctx context.Context, req *proto.PlsResult){
	fmt.Println("someone won the auction")
}
