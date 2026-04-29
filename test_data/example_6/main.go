package main

//go:generate authzed-codegen --schema schema.zed --output ./permissions/

import (
	"context"
	_ "embed"
	"log"
	"os"

	"example_6/permissions"

	authz "github.com/oitnes/authzed-codegen/pkg/authz"
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

	dbAdmin := client.NewUser("db-admin-1")
	outsider := client.NewUser("outsider-1")

	// adminRole is a role that grants db create/read/write permissions.
	// It grants itself the spanner_databases_create permission (self-referential role grant).
	adminRole := client.NewRole("admin-role-1")
	project := client.NewProject("project-1")
	instance := client.NewSpannerInstance("instance-1")
	database := client.NewSpannerDatabase("database-1")

	// Bind the user to the role
	if err := adminRole.CreateBoundUserRelations(ctx, permissions.RoleBoundUserObjects{
		User: []permissions.User{dbAdmin},
	}); err != nil {
		log.Fatal(err)
	}

	// Self-grant: adminRole grants itself spanner_databases_create and spanner_databases_read
	if err := adminRole.CreateSpannerDatabasesCreateRelations(ctx, permissions.RoleSpannerDatabasesCreateObjects{
		Role: []permissions.Role{adminRole},
	}); err != nil {
		log.Fatal(err)
	}
	if err := adminRole.CreateSpannerDatabasesReadRelations(ctx, permissions.RoleSpannerDatabasesReadObjects{
		Role: []permissions.Role{adminRole},
	}); err != nil {
		log.Fatal(err)
	}

	// Grant the role to the project
	if err := project.CreateGrantedRelations(ctx, permissions.ProjectGrantedObjects{
		Role: []permissions.Role{adminRole},
	}); err != nil {
		log.Fatal(err)
	}

	// Link instance to project
	if err := instance.CreateProjectRelations(ctx, permissions.SpannerInstanceProjectObjects{
		Project: []permissions.Project{project},
	}); err != nil {
		log.Fatal(err)
	}

	// Link database to instance
	if err := database.CreateInstanceRelations(ctx, permissions.SpannerDatabaseInstanceObjects{
		SpannerInstance: []permissions.SpannerInstance{instance},
	}); err != nil {
		log.Fatal(err)
	}

	var ok bool
	var err error

	// Verify direct role-level permissions on the role itself
	ok, err = adminRole.CheckCanSpannerDatabasesCreate(ctx, permissions.CheckRoleCanSpannerDatabasesCreateInputs{
		User: []permissions.User{dbAdmin},
	})
	mustTrue(ctx, "role can_spanner_databases_create (self-grant)", ok, err)

	// Verify deep cascade to database: db.create = granted->... + instance->...
	// Use raw engine since subject chain is user → role → project → instance → database
	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permissions.TypeSpannerDatabase, ID: authz.ID(database.ID())},
		permissions.SpannerDatabasePermissionCreate,
		permissions.TypeUser,
		authz.ID(dbAdmin.ID()),
	)
	mustTrue(ctx, "dbAdmin can create database (via role → project → instance → database)", ok, err)

	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permissions.TypeSpannerDatabase, ID: authz.ID(database.ID())},
		permissions.SpannerDatabasePermissionRead,
		permissions.TypeUser,
		authz.ID(dbAdmin.ID()),
	)
	mustTrue(ctx, "dbAdmin can read database", ok, err)

	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permissions.TypeSpannerDatabase, ID: authz.ID(database.ID())},
		permissions.SpannerDatabasePermissionCreate,
		permissions.TypeUser,
		authz.ID(outsider.ID()),
	)
	mustFalse(ctx, "outsider CANNOT create database", ok, err)

	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permissions.TypeSpannerDatabase, ID: authz.ID(database.ID())},
		permissions.SpannerDatabasePermissionRead,
		permissions.TypeUser,
		authz.ID(outsider.ID()),
	)
	mustFalse(ctx, "outsider CANNOT read database", ok, err)

	log.Println("All checks passed — example 6 completed successfully")
}
