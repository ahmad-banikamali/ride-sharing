package main

import (
	"context"
	"fmt"
	"time"
)

func worker(ctx context.Context, name string) {

	for i := 1; ; i++ {
		select {
		case <-ctx.Done():
			fmt.Println(name, "stopping:", ctx.Err())
			return
		default:
		}
		fmt.Println(name, "tick", i)
		time.Sleep(300 * time.Millisecond)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go worker(ctx, "A")

	time.Sleep(5 * time.Second)
	fmt.Println("main: cancel")
	cancel()

	time.Sleep(5 * time.Second)
	fmt.Println("main: exit")
}
