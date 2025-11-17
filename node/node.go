package main

import (
	proto "ITU/grpc"
)

//Leader-based implementation, nodes are replicas
//Like the alpha dawgs

type gangnamStyle struct {
	proto.UnimplementedGangnameStyleServer
	amIsILeader bool           //Which replica is the leader
	wolfPack    []gangnamStyle //List of all replicas

}

func main() {

}

func (gangy *gangnamStyle) selectGangLeader() { // Handles leader failure.
	var isAlphaExist bool // Does leader exist
	for i := 0; i < len(gangy.wolfPack); i++ {
		if gangy.wolfPack[i].amIsILeader { // Goes through the whole list of nodes
			isAlphaExist = true // If the leader is still working the for loop breaks.
			break
		}
	}
	if !isAlphaExist {
		gangy.wolfPack[0].amIsILeader = true // Sets first in list as new leader
	}
}
