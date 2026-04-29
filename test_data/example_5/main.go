package main

//go:generate authzed-codegen --schema schema.zed --output ./permissions/

import (
	"context"
	_ "embed"
	"log"
	"os"

	"example_5/permissions"

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

	sysadmin := client.NewUser("sysadmin-1")
	directOwner := client.NewUser("direct-owner-1")
	outsider := client.NewUser("outsider-1")
	platform := client.NewPlatform("platform-1")
	org := client.NewOrganization("org-1")
	docOwnedByUser := client.NewDocument("doc-user-owned")
	docOwnedByOrg := client.NewDocument("doc-org-owned")

	// Setup: platform administrator
	if err := platform.CreateAdministratorRelations(ctx, permissions.PlatformAdministratorObjects{
		User: []permissions.User{sysadmin},
	}); err != nil {
		log.Fatal(err)
	}

	// Setup: org linked to platform
	if err := org.CreatePlatformRelations(ctx, permissions.OrganizationPlatformObjects{
		Platform: []permissions.Platform{platform},
	}); err != nil {
		log.Fatal(err)
	}

	// Setup: document owned by a direct user
	if err := docOwnedByUser.CreateOwnerRelations(ctx, permissions.DocumentOwnerObjects{
		User: []permissions.User{directOwner},
	}); err != nil {
		log.Fatal(err)
	}

	// Setup: document owned by the org
	if err := docOwnedByOrg.CreateOwnerRelations(ctx, permissions.DocumentOwnerObjects{
		Organization: []permissions.Organization{org},
	}); err != nil {
		log.Fatal(err)
	}

	var ok bool
	var err error

	// platform: super_admin = administrator
	ok, err = platform.CheckSuperAdmin(ctx, permissions.CheckPlatformSuperAdminInputs{User: []permissions.User{sysadmin}})
	mustTrue(ctx, "sysadmin is super_admin on platform", ok, err)
	ok, err = platform.CheckSuperAdmin(ctx, permissions.CheckPlatformSuperAdminInputs{User: []permissions.User{outsider}})
	mustFalse(ctx, "outsider CANNOT super_admin", ok, err)

	// document: admin = owner (direct user owner)
	ok, err = docOwnedByUser.CheckAdmin(ctx, permissions.CheckDocumentAdminInputs{User: []permissions.User{directOwner}})
	mustTrue(ctx, "direct owner can admin their doc", ok, err)
	ok, err = docOwnedByUser.CheckAdmin(ctx, permissions.CheckDocumentAdminInputs{User: []permissions.User{outsider}})
	mustFalse(ctx, "outsider CANNOT admin user-owned doc", ok, err)

	// document: admin = owner->admin (org-owned doc — org.admin = platform->super_admin = sysadmin)
	// The generated CheckAdmin on Document checks Organization as a direct subject type.
	// Use the raw engine to verify the full transitive chain: sysadmin → platform → org → doc.
	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permissions.TypeDocument, ID: authz.ID(docOwnedByOrg.ID())},
		permissions.DocumentPermissionAdmin,
		permissions.TypeUser,
		authz.ID(sysadmin.ID()),
	)
	mustTrue(ctx, "sysadmin can admin org-owned doc (via platform->org->doc)", ok, err)

	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permissions.TypeDocument, ID: authz.ID(docOwnedByOrg.ID())},
		permissions.DocumentPermissionAdmin,
		permissions.TypeUser,
		authz.ID(outsider.ID()),
	)
	mustFalse(ctx, "outsider CANNOT admin org-owned doc", ok, err)

	// org: admin = platform->super_admin (use raw engine for user-level check)
	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permissions.TypeOrganization, ID: authz.ID(org.ID())},
		permissions.OrganizationPermissionAdmin,
		permissions.TypeUser,
		authz.ID(sysadmin.ID()),
	)
	mustTrue(ctx, "sysadmin can admin org (via platform->super_admin)", ok, err)

	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permissions.TypeOrganization, ID: authz.ID(org.ID())},
		permissions.OrganizationPermissionAdmin,
		permissions.TypeUser,
		authz.ID(outsider.ID()),
	)
	mustFalse(ctx, "outsider CANNOT admin org", ok, err)

	log.Println("All checks passed — example 5 completed successfully")
}
