package main

import "log"

func main() {
	log.Println("Starting backend servers...")
	go Server1()
	go Server2()
	select {} //keeps main alive
}
