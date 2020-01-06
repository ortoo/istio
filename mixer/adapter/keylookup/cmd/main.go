package main

import (
	keylookup "istio.io/istio/mixer/adapter/keylookup"
)

func main() {
	server, err := keylookup.NewKeylookup("50051")

	if err != nil {
		panic(err)
	}

	server.Run()
	_ = server.Wait()
}
