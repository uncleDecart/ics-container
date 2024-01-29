package main

import (
	"flag"
	"fmt"

	"github.com/uncleDecart/ics-container/pkg/mock"
)

func main() {
	port := flag.Int("port", 8887, "Port to run mock consumer")

	_, srv := mock.CreateMockConsumer()
	srv.Addr = fmt.Sprintf(":%d", *port)

	fmt.Printf("Serving on %d\n", *port)

	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("Failed to run mock listener %v\n", err)
	}
}
