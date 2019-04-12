package main

import (
	"fmt"
	"os"

	keystonegrpcadapter "istio.io/istio/mixer/adapter/keystonegrpcadapter"
	"istio.io/istio/pkg/log"
)

func main() {
	addr := "9090"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	s, err := keystonegrpcadapter.NewKeystoneGrpcAdapter(addr)
	if err != nil {
		fmt.Printf("unable to start server: %v", err)
		os.Exit(-1)
	}

	shutdown := make(chan error, 1)
	go func() {
		s.Run(shutdown)
	}()
	msg := <-shutdown
	log.Infof("Shutting down with message=%v", msg)
}
