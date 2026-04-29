package main

//go:generate authzed-codegen --schema schema.zed --output ./permissions/

import (
	"context"
	_ "embed"
	"log"
	"os"

	"example_2/permissions"

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

	// Write schema before any relation operations
	if err := engine.EnsureSchema(ctx, schema); err != nil {
		log.Fatalf("failed to write schema: %v", err)
	}
	log.Println("schema written")

	// --- Entities ---
	platform := client.NewPlatform("platform-1")
	anonymousVisitor := client.NewAnonymoususer("anonymous-session-1")
	sysadmin := client.NewUser("sysadmin")
	owner := client.NewUser("owner-user")
	adminUser := client.NewUser("admin-user")
	memberUser := client.NewUser("member-user")
	// bannedAdmin is both an admin and banned — tests the (admin - banned) exclusion
	bannedAdmin := client.NewUser("banned-admin-user")
	outsider := client.NewUser("outsider-user")
	forum := client.NewForum("forum-1")
	post := client.NewPost("post-1")

	// --- Setup: platform ---
	if err := platform.CreateAdministratorRelations(ctx, permissions.PlatformAdministratorObjects{
		User: []permissions.User{sysadmin},
	}); err != nil {
		log.Fatal(err)
	}
	if err := platform.CreateRegisteredUserRelations(ctx, permissions.PlatformRegisteredUserObjects{
		User: []permissions.User{owner, adminUser, memberUser, bannedAdmin},
	}); err != nil {
		log.Fatal(err)
	}
	if err := platform.CreateVisitorRelations(ctx, permissions.PlatformVisitorObjects{
		AnonymoususerWildcard: true,
	}); err != nil {
		log.Fatal(err)
	}

	// --- Setup: forum ---
	if err := forum.CreateGlobalRelations(ctx, permissions.ForumGlobalObjects{
		Platform: []permissions.Platform{platform},
	}); err != nil {
		log.Fatal(err)
	}
	if err := forum.CreateOwnerRelations(ctx, permissions.ForumOwnerObjects{
		User: []permissions.User{owner},
	}); err != nil {
		log.Fatal(err)
	}
	if err := forum.CreateAdminRelations(ctx, permissions.ForumAdminObjects{
		User: []permissions.User{adminUser, bannedAdmin},
	}); err != nil {
		log.Fatal(err)
	}
	if err := forum.CreateMemberRelations(ctx, permissions.ForumMemberObjects{
		User: []permissions.User{memberUser},
	}); err != nil {
		log.Fatal(err)
	}
	if err := forum.CreateBannedRelations(ctx, permissions.ForumBannedObjects{
		User: []permissions.User{bannedAdmin},
	}); err != nil {
		log.Fatal(err)
	}

	// --- Setup: post ---
	if err := post.CreateLocationRelations(ctx, permissions.PostLocationObjects{
		Forum: []permissions.Forum{forum},
	}); err != nil {
		log.Fatal(err)
	}
	// memberUser is author of the post
	if err := post.CreateAuthorRelations(ctx, permissions.PostAuthorObjects{
		User: []permissions.User{memberUser},
	}); err != nil {
		log.Fatal(err)
	}

	// --- Platform: view = visitor + registered_user + administrator ---
	ok, err := platform.CheckView(ctx, permissions.CheckPlatformViewInputs{Anonymoususer: []permissions.Anonymoususer{anonymousVisitor}})
	mustTrue(ctx, "anonymous visitor can view platform", ok, err)
	ok, err = platform.CheckView(ctx, permissions.CheckPlatformViewInputs{User: []permissions.User{outsider}})
	mustFalse(ctx, "outsider user CANNOT view platform", ok, err)

	// --- Platform: create_forum/subscribe_forums/super_admin should deny anonymous ---
	ok, err = platform.CheckCreateForum(ctx, permissions.CheckPlatformCreateForumInputs{Anonymoususer: []permissions.Anonymoususer{anonymousVisitor}})
	mustFalse(ctx, "anonymous visitor CANNOT create forum", ok, err)
	ok, err = platform.CheckSubscribeForums(ctx, permissions.CheckPlatformSubscribeForumsInputs{Anonymoususer: []permissions.Anonymoususer{anonymousVisitor}})
	mustFalse(ctx, "anonymous visitor CANNOT subscribe forums", ok, err)
	ok, err = platform.CheckSuperAdmin(ctx, permissions.CheckPlatformSuperAdminInputs{Anonymoususer: []permissions.Anonymoususer{anonymousVisitor}})
	mustFalse(ctx, "anonymous visitor CANNOT be super_admin", ok, err)

	// --- Forum: make_post = owner + (admin + member - banned) ---
	ok, err = forum.CheckMakePost(ctx, permissions.CheckForumMakePostInputs{User: []permissions.User{owner}})
	mustTrue(ctx, "owner can make_post", ok, err)
	ok, err = forum.CheckMakePost(ctx, permissions.CheckForumMakePostInputs{User: []permissions.User{adminUser}})
	mustTrue(ctx, "admin can make_post", ok, err)
	ok, err = forum.CheckMakePost(ctx, permissions.CheckForumMakePostInputs{User: []permissions.User{memberUser}})
	mustTrue(ctx, "member can make_post", ok, err)
	ok, err = forum.CheckMakePost(ctx, permissions.CheckForumMakePostInputs{User: []permissions.User{bannedAdmin}})
	mustFalse(ctx, "banned admin CANNOT make_post", ok, err)
	ok, err = forum.CheckMakePost(ctx, permissions.CheckForumMakePostInputs{User: []permissions.User{outsider}})
	mustFalse(ctx, "outsider CANNOT make_post", ok, err)

	// --- Forum: view = make_post + global->super_admin ---
	ok, err = forum.CheckView(ctx, permissions.CheckForumViewInputs{User: []permissions.User{owner}})
	mustTrue(ctx, "owner can view forum", ok, err)
	ok, err = forum.CheckView(ctx, permissions.CheckForumViewInputs{User: []permissions.User{adminUser}})
	mustTrue(ctx, "admin can view forum", ok, err)
	ok, err = forum.CheckView(ctx, permissions.CheckForumViewInputs{User: []permissions.User{memberUser}})
	mustTrue(ctx, "member can view forum", ok, err)
	ok, err = forum.CheckView(ctx, permissions.CheckForumViewInputs{User: []permissions.User{sysadmin}})
	mustTrue(ctx, "sysadmin can view forum (via global->super_admin)", ok, err)
	ok, err = forum.CheckView(ctx, permissions.CheckForumViewInputs{User: []permissions.User{bannedAdmin}})
	mustFalse(ctx, "banned admin CANNOT view forum", ok, err)
	ok, err = forum.CheckView(ctx, permissions.CheckForumViewInputs{User: []permissions.User{outsider}})
	mustFalse(ctx, "outsider CANNOT view forum", ok, err)

	// --- Forum: public_view = view + global->view ---
	ok, err = forum.CheckPublicView(ctx, permissions.CheckForumPublicViewInputs{User: []permissions.User{memberUser}})
	mustTrue(ctx, "member can public_view forum", ok, err)
	ok, err = forum.CheckPublicView(ctx, permissions.CheckForumPublicViewInputs{User: []permissions.User{outsider}})
	mustFalse(ctx, "outsider user CANNOT public_view forum", ok, err)
	// Current generated wrapper does not include Anonymoususer input for forum#public_view.
	// Use raw engine check to validate anonymous traversal via global->platform#view(visitor:*).
	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permissions.TypeForum, ID: authz.ID(forum.ID())},
		permissions.ForumPermissionPublicView,
		permissions.TypeAnonymoususer,
		authz.ID(anonymousVisitor.ID()),
	)
	mustTrue(ctx, "anonymous visitor can public_view forum via platform visitor wildcard", ok, err)

	// --- Forum: edit = owner + global->super_admin + (admin - banned) ---
	ok, err = forum.CheckEdit(ctx, permissions.CheckForumEditInputs{User: []permissions.User{owner}})
	mustTrue(ctx, "owner can edit forum", ok, err)
	ok, err = forum.CheckEdit(ctx, permissions.CheckForumEditInputs{User: []permissions.User{sysadmin}})
	mustTrue(ctx, "sysadmin can edit forum", ok, err)
	ok, err = forum.CheckEdit(ctx, permissions.CheckForumEditInputs{User: []permissions.User{adminUser}})
	mustTrue(ctx, "admin can edit forum", ok, err)
	ok, err = forum.CheckEdit(ctx, permissions.CheckForumEditInputs{User: []permissions.User{bannedAdmin}})
	mustFalse(ctx, "banned admin CANNOT edit forum", ok, err)
	ok, err = forum.CheckEdit(ctx, permissions.CheckForumEditInputs{User: []permissions.User{outsider}})
	mustFalse(ctx, "outsider CANNOT edit forum", ok, err)

	// --- Forum: delete = owner + global->super_admin ---
	ok, err = forum.CheckDelete(ctx, permissions.CheckForumDeleteInputs{User: []permissions.User{owner}})
	mustTrue(ctx, "owner can delete forum", ok, err)
	ok, err = forum.CheckDelete(ctx, permissions.CheckForumDeleteInputs{User: []permissions.User{sysadmin}})
	mustTrue(ctx, "sysadmin can delete forum", ok, err)
	ok, err = forum.CheckDelete(ctx, permissions.CheckForumDeleteInputs{User: []permissions.User{adminUser}})
	mustFalse(ctx, "admin CANNOT delete forum", ok, err)
	ok, err = forum.CheckDelete(ctx, permissions.CheckForumDeleteInputs{User: []permissions.User{memberUser}})
	mustFalse(ctx, "member CANNOT delete forum", ok, err)

	// --- Post: edit = author & location->make_post ---
	ok, err = post.CheckEdit(ctx, permissions.CheckPostEditInputs{User: []permissions.User{memberUser}})
	mustTrue(ctx, "memberUser (author + member) can edit post", ok, err)
	ok, err = post.CheckEdit(ctx, permissions.CheckPostEditInputs{User: []permissions.User{owner}})
	mustFalse(ctx, "owner (not author) CANNOT edit post", ok, err)
	ok, err = post.CheckEdit(ctx, permissions.CheckPostEditInputs{User: []permissions.User{outsider}})
	mustFalse(ctx, "outsider CANNOT edit post", ok, err)

	// --- Post: view = location->view ---
	ok, err = post.CheckView(ctx, permissions.CheckPostViewInputs{User: []permissions.User{memberUser}})
	mustTrue(ctx, "member can view post", ok, err)
	ok, err = post.CheckView(ctx, permissions.CheckPostViewInputs{User: []permissions.User{owner}})
	mustTrue(ctx, "owner can view post", ok, err)
	ok, err = post.CheckView(ctx, permissions.CheckPostViewInputs{User: []permissions.User{outsider}})
	mustFalse(ctx, "outsider CANNOT view post", ok, err)
	ok, err = post.CheckView(ctx, permissions.CheckPostViewInputs{User: []permissions.User{bannedAdmin}})
	mustFalse(ctx, "banned admin CANNOT view post", ok, err)
	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permissions.TypePost, ID: authz.ID(post.ID())},
		permissions.PostPermissionView,
		permissions.TypeAnonymoususer,
		authz.ID(anonymousVisitor.ID()),
	)
	mustFalse(ctx, "anonymous visitor CANNOT view post", ok, err)

	log.Println("All checks passed — example 2 completed successfully")
}
