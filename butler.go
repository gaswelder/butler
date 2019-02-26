package main

func main() {
	go trackUpdates()
	go serveBuilds()
	select {}
}
