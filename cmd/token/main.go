package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/auth"
)

func main() {
	userID := flag.String("user", "", "synthetic user ID")
	displayName := flag.String("name", "Demo User", "synthetic display name")
	flag.Parse()
	manager, err := auth.NewManager(os.Getenv("JWT_SECRET"), "go-realtime-messaging", 15*time.Minute)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	token, err := manager.Issue(*userID, *displayName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(token)
}
