//go:build !darwin

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// runApp on non-macOS platforms prints the tablet URL and blocks until
// SIGINT or SIGTERM is received.
func runApp(tabletURL, cfgPath string) {
	fmt.Printf("RemoteBoard running.\n")
	fmt.Printf("  Tablet URL : %s\n", tabletURL)
	fmt.Printf("  Config     : %s\n", cfgPath)
	fmt.Printf("  Press Ctrl-C to quit.\n")
	log.Printf("server: tablet URL is %s", tabletURL)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("remoteboard: exiting")
}
