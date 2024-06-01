package main

import (
	"fmt"
	"runtime"
)

func main() {

	switch runtime.GOOS {
	case "windows":
		//This is where we should build our nestri "nest"
		fmt.Println("You're on Windows!")
	case "darwin":
		//do nothing (probably deploy to AWS, Vast.ai and the rest) plus Linux & Windows
		fmt.Println("You're on macOS!")
	case "linux":
		//This is where we should build our nestri "server"
		fmt.Println("You're on Linux!")
	default:
		fmt.Printf("Unsupported operating system: %s\n", runtime.GOOS)
	}
}
