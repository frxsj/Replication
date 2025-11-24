package main

import (
	proto "ITU/grpc"
	"context"
	"net"
	"log"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//Leader-based implementation, nodes are replicas
//Like the alpha dawgs

type gangnamStyle struct {
	proto.UnimplementedGangnamStyleServer
	portPack []string
	biddersMap map[string] *grpc.ClientConn
	replicas map[string] *grpc.ClientConn
	highestBid map[int64] *grpc.ClientConn
	port string
	amIsILeader bool           //Which replica is the leader
	//highestBid int64
	winner string
}

func main() {
	log.Println("[Server] Starting server...")
	start_server()
	

	wolfPack := []gangnamStyle{}

	for _, wolf := range wolfPack{
		fmt.Println(wolf.amIsILeader)
	}
	
	
	
}

func (gangy *gangnamStyle) forwardToAlpha(req *proto.Request) {
	
}


func (gangy *gangnamStyle) get_available_listener ()(net.Listener){
	//private port range
	var ports []string
	ports = append(ports, ":50000")
	ports = append(ports, ":50001")
	ports = append(ports, ":50002")

	var ret  net.Listener

	for _, port := range ports{
		if (ret == nil){
			listener, err := net.Listen("tcp", port)
			if (err == nil){
				fmt.Println("got port", port)
				gangy.port = port
				ret = listener
				continue
			}
		}

		gangy.replicas[port] = nil

	}

	if ret == nil{
		log.Fatalf("left a lot of crumbs")
	}

	return ret

}

func (gangy *gangnamStyle) connect_nodes() {
	for port, _ := range gangy.replicas {
		conn, err := grpc.NewClient(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to slay")
		}
		gangy.replicas[port] = conn
	}
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

func start_server(){
	node := grpc.NewServer()
	//listener, err := net.Listen("tcp", port)
	gangy := &gangnamStyle{
		replicas: make(map[string]*grpc.ClientConn),
	}
	proto.RegisterGangnamStyleServer(node, gangy)
	listener := gangy.get_available_listener()
	
	go func() {
		if err := node.Serve(listener); err != nil{
			log.Fatalf("womp womp):")
		}
	}()

	gangy.connect_nodes()
}



func (gangy *gangnamStyle) Bid (ctx context.Context, req *proto.Request)(*proto.Response, error){
	//resp := &proto.Response{}

	//hvis amount er valid skal biddet gemmes (hvis det højest)

	//hvis bid ikke er valid skal client få lov til byde igen og fejlbesked

	//hver server har et map (amount, client), hvor kun det højeste bid er gemt
	//sender løbende updates vedrørende højeste bid til clients 
	// broadcastToAll in biddersMap
	/*if error != nil{
		resp {Ack : "Your bet was registered, my drilla😎"}
	} else{
		resp {Ack : "try again my drilla"}
	}
	return resp*/
	return nil, nil
}

func(gangy gangnamStyle) Result (ctx context.Context, req *proto.PlsResult)(outcome *proto.Outcome, err error){
	fmt.Println("someone won the auction")
	return nil, nil
}


