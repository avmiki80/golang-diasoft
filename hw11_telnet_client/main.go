package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	client, closeClient := setupClient()
	defer closeClient()

	ctx := setupSignalContext()

	sendDone, receiveDone := startWorkers(client)

	runMainLoop(ctx, closeClient, sendDone, receiveDone)
}

func setupClient() (TelnetClient, func()) {
	timeout := flag.Duration("timeout", 10*time.Second, "таймаут соединения")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--timeout=<duration>] <host> <port>\n", os.Args[0])
		os.Exit(1)
	}

	host := flag.Arg(0)
	port := flag.Arg(1)
	address := net.JoinHostPort(host, port)

	client := NewTelnetClient(address, *timeout, os.Stdin, os.Stdout)

	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "...Failed to connect to %s: %v\n", address, err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "...Connected to %s\n", address)

	var closeOnce sync.Once
	closeClient := func() {
		closeOnce.Do(func() {
			client.Close()
		})
	}

	return client, closeClient
}

func setupSignalContext() context.Context {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	_ = cancel
	return ctx
}

func startWorkers(client TelnetClient) (chan error, chan error) {
	sendDone := make(chan error, 1)
	receiveDone := make(chan error, 1)

	go func() { sendDone <- runSend(client) }()
	go func() { receiveDone <- runReceive(client) }()

	return sendDone, receiveDone
}

func runSend(client TelnetClient) error {
	if err := client.Send(); err != nil {
		return err
	}
	return io.EOF
}

func runReceive(client TelnetClient) error {
	if err := client.Receive(); err != nil {
		return err
	}
	return io.EOF
}

func runMainLoop(ctx context.Context, closeClient func(), sendDone, receiveDone chan error) {
	for {
		select {
		case <-ctx.Done():
			handleSignalShutdown(closeClient)
			return

		case err := <-sendDone:
			handleSendDone(err, closeClient, receiveDone)
			return

		case err := <-receiveDone:
			handleReceiveDone(err, closeClient, sendDone)
			return
		}
	}
}

func handleSignalShutdown(closeClient func()) {
	fmt.Fprintf(os.Stderr, "\n...Shutting down\n")
	closeClient()
	time.Sleep(100 * time.Millisecond)
}

func handleSendDone(err error, closeClient func(), receiveDone chan error) {
	logError(err, "Send")
	closeClient()
	waitForCompletion(receiveDone, "Receive", 2*time.Second)
}

func handleReceiveDone(err error, closeClient func(), sendDone chan error) {
	logError(err, "Receive")
	closeClient()
	waitForCompletion(sendDone, "Send", 2*time.Second)
}

func logError(err error, operation string) {
	if err != nil && !errors.Is(err, io.EOF) {
		fmt.Fprintf(os.Stderr, "...%s error: %v\n", operation, err)
	}
}

func waitForCompletion(doneChan chan error, operation string, timeout time.Duration) {
	select {
	case err := <-doneChan:
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Fprintf(os.Stderr, "...%s error: %v\n", operation, err)
		}
	case <-time.After(timeout):
		fmt.Fprintf(os.Stderr, "...%s timeout\n", operation)
	}
}
