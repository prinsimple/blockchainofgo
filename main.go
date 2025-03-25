package main

import "github.com/prinsimple/goblock/network"

func main() {
	trLocal := network.NewLocalTransport("local")

	opts := network.ServerOpts{
		Transports: []network.Transport{trLocal},
	}

	server := network.NewServer(opts)

	server.Start()

}
