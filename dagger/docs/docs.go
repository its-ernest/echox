package docs

import (
	"context"
	"os"

	"dagger.io/dagger"
)

func main() {
	ctx := context.Background()
	client, _ := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	defer client.Close()

	// 1. Get the source code
	src := client.Host().Directory(".")

	// 2. Build a container with Go installed
	docs := client.Container().
		From("golang:1.22-bookworm").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		// 3. Install gomarkdoc
		WithExec([]string{"go", "install", "github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest"}).
		// 4. Run the generation
		WithExec([]string{"gomarkdoc", "-o", "{{.Dir}}/README.md", "./..."})

	// 5. Export the changed files back to the host
	_, err := docs.Directory("/src").Export(ctx, ".")
	if err != nil {
		panic(err)
	}
}
