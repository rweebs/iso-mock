package main

// GitCommit will be used as service version
var GitCommit string

func main() {
	command := NewCommand("Mocking Biller")
	command.Run()
}
