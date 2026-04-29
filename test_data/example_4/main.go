package main

//go:generate authzed-codegen --schema schema.zed --output ./permissions/

import (
	"context"
	_ "embed"
	"log"
	"os"

	"example_4/permissions"

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

	member := client.NewUser("member-1")
	outsider := client.NewUser("outsider-1")
	org := client.NewOrganization("org-1")
	ent := client.NewEntitlement("entitlement-1")
	feature := client.NewFeature("feature-1")

	if err := org.CreateMemberRelations(ctx, permissions.OrganizationMemberObjects{
		User: []permissions.User{member},
	}); err != nil {
		log.Fatal(err)
	}
	if err := ent.CreateOrgRelations(ctx, permissions.EntitlementOrgObjects{
		Organization: []permissions.Organization{org},
	}); err != nil {
		log.Fatal(err)
	}
	if err := feature.CreateAssociatedEntitlementRelations(ctx, permissions.FeatureAssociatedEntitlementObjects{
		Entitlement: []permissions.Entitlement{ent},
	}); err != nil {
		log.Fatal(err)
	}

	// feature.access = associated_entitlement->subscribed_member, subscribed_member = org->member
	// Subject types for CheckAccess are Entitlement (direct relation type on feature), so use raw engine
	// to check user-level access through the full arrow chain.
	ok, err := engine.CheckPermission(
		ctx,
		authz.Resource{Type: permissions.TypeFeature, ID: authz.ID(feature.ID())},
		permissions.FeaturePermissionAccess,
		permissions.TypeUser,
		authz.ID(member.ID()),
	)
	mustTrue(ctx, "org member can access feature", ok, err)

	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permissions.TypeFeature, ID: authz.ID(feature.ID())},
		permissions.FeaturePermissionAccess,
		permissions.TypeUser,
		authz.ID(outsider.ID()),
	)
	mustFalse(ctx, "outsider CANNOT access feature", ok, err)

	log.Println("All checks passed — example 4 completed successfully")
}
