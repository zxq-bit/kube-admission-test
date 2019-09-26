package main

import (
	"log"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/server"
)

func main() {
	s, e := server.NewServer()
	if e != nil {
		log.Fatalf("NewServer failed, %v", e)
	}
	log.Printf("Starting admission server")

	if e = s.Run(); e != nil {
		log.Fatalf("Run failed, %v", e)
	}
}
