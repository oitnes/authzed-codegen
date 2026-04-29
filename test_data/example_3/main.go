package main

//go:generate authzed-codegen --schema schema.zed --output ./permissions/

import (
	"context"
	_ "embed"
	"log"
	"os"

	"example_3/permissions"

	spicedbengine "github.com/oitnes/authzed-codegen/pkg/authz/spicedb"
)

//go:embed schema.zed
var schema string

func newEngine() *spicedbengine.Engine {
	endpoint := os.Getenv("SPICEDB_ENDPOINT")
	token := os.Getenv("SPICEDB_TOKEN")
	if endpoint == "" || token == "" {
		log.Fatal("SPICEDB_ENDPOINT and SPICEDB_TOKEN environment variables are required")
	}
	engine, err := spicedbengine.NewEngine(endpoint, token)
	if err != nil {
		log.Fatalf("failed to connect to SpiceDB: %v", err)
	}
	log.Printf("connected to SpiceDB at %s", endpoint)
	return engine
}

func mustTrue(ctx context.Context, label string, got bool, err error) {
	if err != nil {
		log.Fatalf("FAIL [%s]: unexpected error: %v", label, err)
	}
	if !got {
		log.Fatalf("FAIL [%s]: expected true, got false", label)
	}
	log.Printf("PASS [%s]", label)
}

func mustFalse(ctx context.Context, label string, got bool, err error) {
	if err != nil {
		log.Fatalf("FAIL [%s]: unexpected error: %v", label, err)
	}
	if got {
		log.Fatalf("FAIL [%s]: expected false, got true", label)
	}
	log.Printf("PASS [%s]", label)
}

func main() {
	ctx := context.Background()
	engine := newEngine()
	client := permissions.NewClient(engine)

	if err := engine.EnsureSchema(ctx, schema); err != nil {
		log.Fatalf("failed to write schema: %v", err)
	}
	log.Println("schema written")

	writer := client.NewUser("writer-1")
	reader := client.NewUser("reader-1")
	outsider := client.NewUser("outsider-1")
	doc := client.NewDocument("doc-1")

	if err := doc.CreateWriterRelations(ctx, permissions.DocumentWriterObjects{
		User: []permissions.User{writer},
	}); err != nil {
		log.Fatal(err)
	}
	if err := doc.CreateReaderRelations(ctx, permissions.DocumentReaderObjects{
		User: []permissions.User{reader},
	}); err != nil {
		log.Fatal(err)
	}

	var ok bool
	var err error

	// edit = writer
	ok, err = doc.CheckEdit(ctx, permissions.CheckDocumentEditInputs{User: []permissions.User{writer}})
	mustTrue(ctx, "writer can edit", ok, err)
	ok, err = doc.CheckEdit(ctx, permissions.CheckDocumentEditInputs{User: []permissions.User{reader}})
	mustFalse(ctx, "reader CANNOT edit", ok, err)
	ok, err = doc.CheckEdit(ctx, permissions.CheckDocumentEditInputs{User: []permissions.User{outsider}})
	mustFalse(ctx, "outsider CANNOT edit", ok, err)

	// view = reader + edit
	ok, err = doc.CheckView(ctx, permissions.CheckDocumentViewInputs{User: []permissions.User{writer}})
	mustTrue(ctx, "writer can view (via edit)", ok, err)
	ok, err = doc.CheckView(ctx, permissions.CheckDocumentViewInputs{User: []permissions.User{reader}})
	mustTrue(ctx, "reader can view", ok, err)
	ok, err = doc.CheckView(ctx, permissions.CheckDocumentViewInputs{User: []permissions.User{outsider}})
	mustFalse(ctx, "outsider CANNOT view", ok, err)

	log.Println("All checks passed — example 3 completed successfully")
}
