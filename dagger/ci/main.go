package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

func main() {
	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// 1. Get the project source
	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{".git", "bin", "obj"},
	})

	// 2. Define the Go execution environment
	gopher := client.Container().
		From("golang:1.25-bookworm").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		// Cache the Go modules to speed up subsequent runs
		WithMountedCache("/go/pkg/mod", client.CacheVolume("go-mod"))

	// 3. Run the Tests
	fmt.Println("Running unit tests...")
	_, err = gopher.
		//WithExec([]string{"go", "test", "-v", "-race", "./..."}).
		WithExec([]string{"go", "test", "-v", "-race", "./..."}).
		Sync(ctx)

	if err != nil {
		fmt.Printf("Tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All tests passed!")
}
