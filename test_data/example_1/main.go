package main

//go:generate authzed-codegen --schema schema.zed --output ./permitions/

import (
	"context"
	_ "embed"
	"log"
	"os"

	"example_1/permitions"

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

	if err := engine.EnsureSchema(ctx, schema); err != nil {
		log.Fatalf("failed to write schema: %v", err)
	}
	log.Println("schema written")

	client := permitions.NewClient(engine)

	// --- Entities: bookingsvc ---
	brandAdminUser := client.NewBookingsvcUser("brand-admin-user")
	outsiderUser := client.NewBookingsvcUser("outsider-user")

	brand := client.NewBookingsvcBrand("brand-1")
	managerEmployee := client.NewBookingsvcEmployee("manager-emp-1")
	brandOnlyEmployee := client.NewBookingsvcEmployee("brand-only-emp-1")
	outsiderEmployee := client.NewBookingsvcEmployee("outsider-emp-1")
	ownerEmployee := client.NewBookingsvcEmployee("owner-emp-1")
	creatorCustomer := client.NewBookingsvcCustomer("creator-customer-1")
	booking := client.NewBookingsvcBooking("booking-1")

	// --- Entities: menusvc ---
	companyAdminUser := client.NewMenusvcUser("company-admin-1")
	companyManagerUser := client.NewMenusvcUser("company-manager-1")
	companyEmployeeUser := client.NewMenusvcUser("company-employee-1")
	outsiderMenuUser := client.NewMenusvcUser("outsider-menu-user")
	menuCustomer := client.NewMenusvcCustomer("menu-customer-1")

	company := client.NewMenusvcCompany("company-1")
	outsiderCompany := client.NewMenusvcCompany("outsider-company-1")
	menuBooking := client.NewMenusvcBooking("menu-booking-1")
	menuOrder := client.NewMenusvcOrder("menu-order-1")
	menuTable := client.NewMenusvcTable("menu-table-1")
	menuPricelist := client.NewMenusvcPricelist("menu-pricelist-1")
	menuSetting := client.NewMenusvcSetting("menu-setting-1")

	// --- Setup: bookingsvc/brand ---
	if err := brand.CreateAdminRelations(ctx, permitions.BookingsvcBrandAdminObjects{
		BookingsvcUser: []permitions.BookingsvcUser{brandAdminUser},
	}); err != nil {
		log.Fatal(err)
	}
	if err := brand.CreateManagerRelations(ctx, permitions.BookingsvcBrandManagerObjects{
		BookingsvcEmployee: []permitions.BookingsvcEmployee{managerEmployee},
	}); err != nil {
		log.Fatal(err)
	}
	if err := brand.CreateEmployeeRelations(ctx, permitions.BookingsvcBrandEmployeeObjects{
		BookingsvcEmployee: []permitions.BookingsvcEmployee{brandOnlyEmployee},
	}); err != nil {
		log.Fatal(err)
	}

	// managerEmployee belongs to brand so it inherits brand->manage
	if err := managerEmployee.CreateBelongsBrandRelations(ctx, permitions.BookingsvcEmployeeBelongsBrandObjects{
		BookingsvcBrand: []permitions.BookingsvcBrand{brand},
	}); err != nil {
		log.Fatal(err)
	}
	// ownerEmployee has brandAdminUser as its linked account
	if err := ownerEmployee.CreateAccountRelations(ctx, permitions.BookingsvcEmployeeAccountObjects{
		BookingsvcUser: []permitions.BookingsvcUser{brandAdminUser},
	}); err != nil {
		log.Fatal(err)
	}

	// --- Setup: bookingsvc/booking ---
	if err := booking.CreateOwnerRelations(ctx, permitions.BookingsvcBookingOwnerObjects{
		BookingsvcEmployee: []permitions.BookingsvcEmployee{ownerEmployee},
	}); err != nil {
		log.Fatal(err)
	}
	if err := booking.CreateCreatorRelations(ctx, permitions.BookingsvcBookingCreatorObjects{
		BookingsvcCustomer: []permitions.BookingsvcCustomer{creatorCustomer},
	}); err != nil {
		log.Fatal(err)
	}

	// --- bookingsvc/brand: manage = manager + admin ---
	ok, err := brand.CheckManage(ctx, permitions.CheckBookingsvcBrandManageInputs{
		BookingsvcUser: []permitions.BookingsvcUser{brandAdminUser},
	})
	mustTrue(ctx, "brand admin user can manage brand", ok, err)

	ok, err = brand.CheckManage(ctx, permitions.CheckBookingsvcBrandManageInputs{
		BookingsvcEmployee: []permitions.BookingsvcEmployee{managerEmployee},
	})
	mustTrue(ctx, "manager employee can manage brand (via manager relation)", ok, err)

	ok, err = brand.CheckManage(ctx, permitions.CheckBookingsvcBrandManageInputs{
		BookingsvcEmployee: []permitions.BookingsvcEmployee{brandOnlyEmployee},
	})
	mustFalse(ctx, "brand-only employee CANNOT manage brand", ok, err)

	ok, err = brand.CheckManage(ctx, permitions.CheckBookingsvcBrandManageInputs{
		BookingsvcUser: []permitions.BookingsvcUser{outsiderUser},
	})
	mustFalse(ctx, "outsider user CANNOT manage brand", ok, err)

	// --- bookingsvc/brand: create_booking = manage + employee ---
	ok, err = brand.CheckCreateBooking(ctx, permitions.CheckBookingsvcBrandCreateBookingInputs{
		BookingsvcUser: []permitions.BookingsvcUser{brandAdminUser},
	})
	mustTrue(ctx, "admin user can create_booking on brand (via manage)", ok, err)

	ok, err = brand.CheckCreateBooking(ctx, permitions.CheckBookingsvcBrandCreateBookingInputs{
		BookingsvcEmployee: []permitions.BookingsvcEmployee{managerEmployee},
	})
	mustTrue(ctx, "manager employee can create_booking on brand", ok, err)

	ok, err = brand.CheckCreateBooking(ctx, permitions.CheckBookingsvcBrandCreateBookingInputs{
		BookingsvcEmployee: []permitions.BookingsvcEmployee{brandOnlyEmployee},
	})
	mustTrue(ctx, "brand-only employee can create_booking on brand", ok, err)

	ok, err = brand.CheckCreateBooking(ctx, permitions.CheckBookingsvcBrandCreateBookingInputs{
		BookingsvcEmployee: []permitions.BookingsvcEmployee{outsiderEmployee},
	})
	mustFalse(ctx, "outsider employee CANNOT create_booking on brand", ok, err)

	// --- bookingsvc/employee: manage = account + belongs_brand->manage ---
	ok, err = ownerEmployee.CheckManage(ctx, permitions.CheckBookingsvcEmployeeManageInputs{
		BookingsvcUser: []permitions.BookingsvcUser{brandAdminUser},
	})
	mustTrue(ctx, "account user can manage employee (via account relation)", ok, err)

	// belongs_brand->manage: subjects are the users/employees who have manage on the brand.
	// The generated wrapper exposes BookingsvcBrand as subject type (stops at the arrow target),
	// but the actual actors are BookingsvcUser/Employee. Use raw engine to verify the transitive path.
	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permitions.TypeBookingsvcEmployee, ID: authz.ID(managerEmployee.ID())},
		permitions.BookingsvcEmployeePermissionManage,
		permitions.TypeBookingsvcUser,
		authz.ID(brandAdminUser.ID()),
	)
	mustTrue(ctx, "brand admin user can manage managerEmployee (via belongs_brand->manage->admin)", ok, err)

	ok, err = ownerEmployee.CheckManage(ctx, permitions.CheckBookingsvcEmployeeManageInputs{
		BookingsvcUser: []permitions.BookingsvcUser{outsiderUser},
	})
	mustFalse(ctx, "outsider user CANNOT manage employee", ok, err)

	// --- bookingsvc/employee: view = manage + viewer ---
	ok, err = ownerEmployee.CheckView(ctx, permitions.CheckBookingsvcEmployeeViewInputs{
		BookingsvcUser: []permitions.BookingsvcUser{brandAdminUser},
	})
	mustTrue(ctx, "account user can view employee (via manage)", ok, err)

	ok, err = ownerEmployee.CheckView(ctx, permitions.CheckBookingsvcEmployeeViewInputs{
		BookingsvcUser: []permitions.BookingsvcUser{outsiderUser},
	})
	mustFalse(ctx, "outsider user CANNOT view employee before wildcard", ok, err)

	// Set viewer:* so any bookingsvc/user can view the employee
	if err = ownerEmployee.CreateViewerRelations(ctx, permitions.BookingsvcEmployeeViewerObjects{
		BookingsvcUserWildcard: true,
	}); err != nil {
		log.Fatal(err)
	}
	ok, err = ownerEmployee.CheckView(ctx, permitions.CheckBookingsvcEmployeeViewInputs{
		BookingsvcUser: []permitions.BookingsvcUser{outsiderUser},
	})
	mustTrue(ctx, "outsider user CAN view employee after viewer wildcard", ok, err)

	// --- bookingsvc/booking: write = creator + owner + owner->manage + creator->manage ---
	ok, err = booking.CheckWrite(ctx, permitions.CheckBookingsvcBookingWriteInputs{
		BookingsvcEmployee: []permitions.BookingsvcEmployee{ownerEmployee},
	})
	mustTrue(ctx, "owner employee can write booking (via owner relation)", ok, err)

	ok, err = booking.CheckWrite(ctx, permitions.CheckBookingsvcBookingWriteInputs{
		BookingsvcCustomer: []permitions.BookingsvcCustomer{creatorCustomer},
	})
	mustTrue(ctx, "creator customer can write booking (via creator relation)", ok, err)

	ok, err = booking.CheckWrite(ctx, permitions.CheckBookingsvcBookingWriteInputs{
		BookingsvcEmployee: []permitions.BookingsvcEmployee{outsiderEmployee},
	})
	mustFalse(ctx, "outsider employee CANNOT write booking", ok, err)

	// --- bookingsvc/booking: change_owner = creator + creator->manage ---
	ok, err = booking.CheckChangeOwner(ctx, permitions.CheckBookingsvcBookingChangeOwnerInputs{
		BookingsvcCustomer: []permitions.BookingsvcCustomer{creatorCustomer},
	})
	mustTrue(ctx, "creator customer can change_owner of booking", ok, err)

	ok, err = booking.CheckChangeOwner(ctx, permitions.CheckBookingsvcBookingChangeOwnerInputs{
		BookingsvcEmployee: []permitions.BookingsvcEmployee{ownerEmployee},
	})
	mustFalse(ctx, "owner employee CANNOT change_owner (owner is not in change_owner)", ok, err)

	ok, err = booking.CheckChangeOwner(ctx, permitions.CheckBookingsvcBookingChangeOwnerInputs{
		BookingsvcEmployee: []permitions.BookingsvcEmployee{outsiderEmployee},
	})
	mustFalse(ctx, "outsider employee CANNOT change_owner", ok, err)

	// --- Setup: menusvc/company ---
	if err = company.CreateAdminRelations(ctx, permitions.MenusvcCompanyAdminObjects{
		MenusvcUser: []permitions.MenusvcUser{companyAdminUser},
	}); err != nil {
		log.Fatal(err)
	}
	if err = company.CreateManagerRelations(ctx, permitions.MenusvcCompanyManagerObjects{
		MenusvcUser: []permitions.MenusvcUser{companyManagerUser},
	}); err != nil {
		log.Fatal(err)
	}
	if err = company.CreateEmployeeRelations(ctx, permitions.MenusvcCompanyEmployeeObjects{
		MenusvcUser: []permitions.MenusvcUser{companyEmployeeUser},
	}); err != nil {
		log.Fatal(err)
	}

	// companyAdminUser belongs to company (enables manage via belongs_company->manage)
	if err = companyAdminUser.CreateBelongsCompanyRelations(ctx, permitions.MenusvcUserBelongsCompanyObjects{
		MenusvcCompany: []permitions.MenusvcCompany{company},
	}); err != nil {
		log.Fatal(err)
	}

	// --- menusvc/company: manage = admin + manager ---
	ok, err = company.CheckManage(ctx, permitions.CheckMenusvcCompanyManageInputs{
		MenusvcUser: []permitions.MenusvcUser{companyAdminUser},
	})
	mustTrue(ctx, "company admin can manage company", ok, err)

	ok, err = company.CheckManage(ctx, permitions.CheckMenusvcCompanyManageInputs{
		MenusvcUser: []permitions.MenusvcUser{companyManagerUser},
	})
	mustTrue(ctx, "company manager can manage company", ok, err)

	ok, err = company.CheckManage(ctx, permitions.CheckMenusvcCompanyManageInputs{
		MenusvcUser: []permitions.MenusvcUser{companyEmployeeUser},
	})
	mustFalse(ctx, "company employee CANNOT manage company", ok, err)

	ok, err = company.CheckManage(ctx, permitions.CheckMenusvcCompanyManageInputs{
		MenusvcUser: []permitions.MenusvcUser{outsiderMenuUser},
	})
	mustFalse(ctx, "outsider menu user CANNOT manage company", ok, err)

	// --- menusvc/company: create_booking = manage + employee ---
	ok, err = company.CheckCreateBooking(ctx, permitions.CheckMenusvcCompanyCreateBookingInputs{
		MenusvcUser: []permitions.MenusvcUser{companyAdminUser},
	})
	mustTrue(ctx, "company admin can create_booking (via manage)", ok, err)

	ok, err = company.CheckCreateBooking(ctx, permitions.CheckMenusvcCompanyCreateBookingInputs{
		MenusvcUser: []permitions.MenusvcUser{companyEmployeeUser},
	})
	mustTrue(ctx, "company employee can create_booking", ok, err)

	ok, err = company.CheckCreateBooking(ctx, permitions.CheckMenusvcCompanyCreateBookingInputs{
		MenusvcUser: []permitions.MenusvcUser{outsiderMenuUser},
	})
	mustFalse(ctx, "outsider menu user CANNOT create_booking", ok, err)

	// --- menusvc/company: create_order = manage + employee ---
	ok, err = company.CheckCreateOrder(ctx, permitions.CheckMenusvcCompanyCreateOrderInputs{
		MenusvcUser: []permitions.MenusvcUser{companyManagerUser},
	})
	mustTrue(ctx, "company manager can create_order (via manage)", ok, err)

	ok, err = company.CheckCreateOrder(ctx, permitions.CheckMenusvcCompanyCreateOrderInputs{
		MenusvcUser: []permitions.MenusvcUser{companyEmployeeUser},
	})
	mustTrue(ctx, "company employee can create_order", ok, err)

	ok, err = company.CheckCreateOrder(ctx, permitions.CheckMenusvcCompanyCreateOrderInputs{
		MenusvcUser: []permitions.MenusvcUser{outsiderMenuUser},
	})
	mustFalse(ctx, "outsider menu user CANNOT create_order", ok, err)

	// --- menusvc/user: manage = belongs_company->manage ---
	// The generated wrapper exposes MenusvcCompany as subject type for this arrow permission.
	// Use the raw engine to check with the actual user-level subjects.
	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permitions.TypeMenusvcUser, ID: authz.ID(companyAdminUser.ID())},
		permitions.MenusvcUserPermissionManage,
		permitions.TypeMenusvcUser,
		authz.ID(companyAdminUser.ID()),
	)
	mustTrue(ctx, "companyAdminUser can manage themselves (belongs to company they admin)", ok, err)

	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permitions.TypeMenusvcUser, ID: authz.ID(companyAdminUser.ID())},
		permitions.MenusvcUserPermissionManage,
		permitions.TypeMenusvcUser,
		authz.ID(outsiderMenuUser.ID()),
	)
	mustFalse(ctx, "outsider menu user CANNOT manage companyAdminUser", ok, err)

	// --- Setup: menusvc/booking ---
	if err = menuBooking.CreateOwnerRelations(ctx, permitions.MenusvcBookingOwnerObjects{
		MenusvcCompany: []permitions.MenusvcCompany{company},
	}); err != nil {
		log.Fatal(err)
	}
	if err = menuBooking.CreateCreatorRelations(ctx, permitions.MenusvcBookingCreatorObjects{
		MenusvcUser:     []permitions.MenusvcUser{companyEmployeeUser},
		MenusvcCustomer: []permitions.MenusvcCustomer{menuCustomer},
	}); err != nil {
		log.Fatal(err)
	}

	// --- menusvc/booking: write = creator + creator->manage + owner->manage ---
	ok, err = menuBooking.CheckWrite(ctx, permitions.CheckMenusvcBookingWriteInputs{
		MenusvcUser: []permitions.MenusvcUser{companyEmployeeUser},
	})
	mustTrue(ctx, "creator user can write menu booking", ok, err)

	ok, err = menuBooking.CheckWrite(ctx, permitions.CheckMenusvcBookingWriteInputs{
		MenusvcCustomer: []permitions.MenusvcCustomer{menuCustomer},
	})
	mustTrue(ctx, "creator customer can write menu booking", ok, err)

	ok, err = menuBooking.CheckWrite(ctx, permitions.CheckMenusvcBookingWriteInputs{
		MenusvcUser: []permitions.MenusvcUser{outsiderMenuUser},
	})
	mustFalse(ctx, "outsider menu user CANNOT write menu booking", ok, err)

	// Company admin can write via owner->manage — use raw engine (transitive arrow)
	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permitions.TypeMenusvcBooking, ID: authz.ID(menuBooking.ID())},
		permitions.MenusvcBookingPermissionWrite,
		permitions.TypeMenusvcUser,
		authz.ID(companyAdminUser.ID()),
	)
	mustTrue(ctx, "company admin can write menu booking (via owner->manage)", ok, err)

	// --- Setup: menusvc/order ---
	if err = menuOrder.CreateBelongsCompanyRelations(ctx, permitions.MenusvcOrderBelongsCompanyObjects{
		MenusvcCompany: []permitions.MenusvcCompany{company},
	}); err != nil {
		log.Fatal(err)
	}
	if err = menuOrder.CreateCreatorRelations(ctx, permitions.MenusvcOrderCreatorObjects{
		MenusvcUser: []permitions.MenusvcUser{companyEmployeeUser},
	}); err != nil {
		log.Fatal(err)
	}

	// --- menusvc/order: write = creator + creator->manage + belongs_company->manage ---
	ok, err = menuOrder.CheckWrite(ctx, permitions.CheckMenusvcOrderWriteInputs{
		MenusvcUser: []permitions.MenusvcUser{companyEmployeeUser},
	})
	mustTrue(ctx, "creator user can write menu order", ok, err)

	ok, err = menuOrder.CheckWrite(ctx, permitions.CheckMenusvcOrderWriteInputs{
		MenusvcUser: []permitions.MenusvcUser{outsiderMenuUser},
	})
	mustFalse(ctx, "outsider menu user CANNOT write menu order", ok, err)

	// Company admin can write via belongs_company->manage — use raw engine (transitive arrow)
	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permitions.TypeMenusvcOrder, ID: authz.ID(menuOrder.ID())},
		permitions.MenusvcOrderPermissionWrite,
		permitions.TypeMenusvcUser,
		authz.ID(companyAdminUser.ID()),
	)
	mustTrue(ctx, "company admin can write menu order (via belongs_company->manage)", ok, err)

	// --- Setup: menusvc/table, pricelist, setting owned by company ---
	if err = menuTable.CreateOwnerRelations(ctx, permitions.MenusvcTableOwnerObjects{
		MenusvcCompany: []permitions.MenusvcCompany{company},
	}); err != nil {
		log.Fatal(err)
	}
	if err = menuPricelist.CreateOwnerRelations(ctx, permitions.MenusvcPricelistOwnerObjects{
		MenusvcCompany: []permitions.MenusvcCompany{company},
	}); err != nil {
		log.Fatal(err)
	}
	if err = menuSetting.CreateOwnerRelations(ctx, permitions.MenusvcSettingOwnerObjects{
		MenusvcCompany: []permitions.MenusvcCompany{company},
	}); err != nil {
		log.Fatal(err)
	}

	// --- menusvc/table: write = owner->manage (use raw engine — transitive arrow) ---
	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permitions.TypeMenusvcTable, ID: authz.ID(menuTable.ID())},
		permitions.MenusvcTablePermissionWrite,
		permitions.TypeMenusvcUser,
		authz.ID(companyAdminUser.ID()),
	)
	mustTrue(ctx, "company admin can write table (via owner->manage)", ok, err)

	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permitions.TypeMenusvcTable, ID: authz.ID(menuTable.ID())},
		permitions.MenusvcTablePermissionWrite,
		permitions.TypeMenusvcUser,
		authz.ID(companyEmployeeUser.ID()),
	)
	mustFalse(ctx, "company employee CANNOT write table (not manager/admin)", ok, err)

	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permitions.TypeMenusvcTable, ID: authz.ID(menuTable.ID())},
		permitions.MenusvcTablePermissionWrite,
		permitions.TypeMenusvcUser,
		authz.ID(outsiderMenuUser.ID()),
	)
	mustFalse(ctx, "outsider menu user CANNOT write table", ok, err)

	// --- menusvc/table: no write for outsider company ---
	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permitions.TypeMenusvcTable, ID: authz.ID(menuTable.ID())},
		permitions.MenusvcTablePermissionWrite,
		permitions.TypeMenusvcCompany,
		authz.ID(outsiderCompany.ID()),
	)
	mustFalse(ctx, "outsider company CANNOT write table", ok, err)

	// --- menusvc/pricelist: write = owner->manage ---
	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permitions.TypeMenusvcPricelist, ID: authz.ID(menuPricelist.ID())},
		permitions.MenusvcPricelistPermissionWrite,
		permitions.TypeMenusvcUser,
		authz.ID(companyManagerUser.ID()),
	)
	mustTrue(ctx, "company manager can write pricelist (via owner->manage)", ok, err)

	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permitions.TypeMenusvcPricelist, ID: authz.ID(menuPricelist.ID())},
		permitions.MenusvcPricelistPermissionWrite,
		permitions.TypeMenusvcUser,
		authz.ID(outsiderMenuUser.ID()),
	)
	mustFalse(ctx, "outsider menu user CANNOT write pricelist", ok, err)

	// --- menusvc/setting: write = owner->manage ---
	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permitions.TypeMenusvcSetting, ID: authz.ID(menuSetting.ID())},
		permitions.MenusvcSettingPermissionWrite,
		permitions.TypeMenusvcUser,
		authz.ID(companyAdminUser.ID()),
	)
	mustTrue(ctx, "company admin can write setting (via owner->manage)", ok, err)

	ok, err = engine.CheckPermission(
		ctx,
		authz.Resource{Type: permitions.TypeMenusvcSetting, ID: authz.ID(menuSetting.ID())},
		permitions.MenusvcSettingPermissionWrite,
		permitions.TypeMenusvcUser,
		authz.ID(outsiderMenuUser.ID()),
	)
	mustFalse(ctx, "outsider menu user CANNOT write setting", ok, err)

	log.Println("All checks passed — example 1 completed successfully")
}
