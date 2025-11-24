package main

import (
	proto "ITU/grpc"
	"context"
	"net"
	"log"
	"fmt"
	"os"
	"bufio"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//Leader-based implementation, nodes are replicas
//Like the alpha dawgs



type gangnamStyle struct {
	proto.UnimplementedGangnamStyleServer
	conn *grpc.ClientConn
	biddersMap map[string] *grpc.ClientConn
	replicas map[string] *gangnamStyle
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
	ports = append(ports, ":50002")
	ports = append(ports, ":50003")
	ports = append(ports, ":50004")

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
	fmt.Println("connecting the reaplicas...")
	for port, _ := range gangy.replicas {
		conn, err := grpc.NewClient(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to slay")
		}
		gang := &gangnamStyle{
			conn: conn,
			port: port,
		}

		gangy.replicas[port] = gang
	}
}


func (gangy *gangnamStyle)selectGangLeader() { // Handles leader failure.
	fmt.Println("banke banke på")
	var isAlphaExist bool // Does leader exist


	for _, replica := range gangy.replicas{
		if(replica.amIsILeader){
			isAlphaExist = true
			break
		}
	}

	if (!isAlphaExist){
		for _, replica := range gangy.replicas{
			replica.amIsILeader = true
			fmt.Println("The new alpha has been chosen🐺 on port", replica.port)
			break//dance
		}
	}

}

func start_server(){
	node := grpc.NewServer()
	//listener, err := net.Listen("tcp", port)
	gangy := &gangnamStyle{
		replicas: make(map[string]*gangnamStyle),
	}
	proto.RegisterGangnamStyleServer(node, gangy)
	listener := gangy.get_available_listener()
	
	go func() {
		if err := node.Serve(listener); err != nil{
			log.Fatalf("womp womp):")
		}
	}()
	fmt.Println("Once all replicas have been started type 'arla cultura 4eva<3<3'")
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan(){
		if(scanner.Text() == "arla cultura 4eva<3<3"){
			gangy.connect_nodes()
			gangy.selectGangLeader()
			break
		}
	}

	time.Sleep(20* time.Second)

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

func(gangy *gangnamStyle) Result (ctx context.Context, req *proto.PlsResult)(*proto.Outcome, error){
	fmt.Println("someone won the auction")
	return nil, nil
}

func (gangy *gangnamStyle) Bully (ctx context.Context, b_req *proto.BullyRequest)(*proto.BullyResponse, error){
	fmt.Println("The bullying has begun!")

	higherReplicas := []*gangnamStyle{}

	for _, replica := range gangy.replicas{ //tilføjer alle replicas der har en højere id (bruger bare port) til et slice
		if(replica.port > gangy.port){
			higherReplicas = append(higherReplicas, replica) 
		}
	}

	if (len(higherReplicas) == 0){ //hvis slice er tomt betyder det at replica på denne process må have det højeste id (port) og bliver derfor leader
		gangy.amIsILeader = true
	}

	return nil,nil
}


