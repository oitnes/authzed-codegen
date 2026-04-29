package authz

import "context"

// Type represents a SpiceDB object type name.
type Type string

// ID represents a SpiceDB object identifier.
type ID string

// Permission represents a permission name in SpiceDB.
type Permission string

// Relation represents a relation name in SpiceDB.
type Relation string

// Resource identifies a specific object in SpiceDB.
type Resource struct {
	Type Type
	ID   ID
}

// PermissionCheck represents a single permission check in a bulk operation.
type PermissionCheck struct {
	Resource    Resource
	Permission  Permission
	SubjectType Type
	SubjectID   ID
}

// RelationshipObject represents a relationship for bulk import/export operations.
type RelationshipObject struct {
	Resource    Resource
	Relation    Relation
	SubjectType Type
	SubjectID   ID
}

// RelationshipFilter specifies criteria for exporting relationships.
type RelationshipFilter struct {
	ResourceType string
	ResourceID   string
	Relation     string
	SubjectType  string
}

// Engine defines the interface for SpiceDB authorization operations.
// Generated code calls these methods via constructor-injected instances.
type Engine interface {
	// Core relation operations
	CreateRelations(ctx context.Context, resource Resource, relation Relation, subjectType Type, subjectIDs []ID) error
	ReadRelations(ctx context.Context, resource Resource, relation Relation, subjectType Type) ([]ID, error)
	DeleteRelations(ctx context.Context, resource Resource, relation Relation, subjectType Type, subjectIDs []ID) error

	// Core permission operations
	CheckPermission(ctx context.Context, resource Resource, permission Permission, subjectType Type, subjectID ID) (bool, error)
	LookupResources(ctx context.Context, resourceType Type, permission Permission, subjectType Type, subjectID ID) ([]ID, error)
	LookupSubjects(ctx context.Context, resource Resource, permission Permission, subjectType Type) ([]ID, error)

	// Bulk operations
	CheckBulkPermission(ctx context.Context, checks []PermissionCheck) ([]bool, error)
	ExportBulkRelationships(ctx context.Context, filter RelationshipFilter) ([]RelationshipObject, error)
	ImportBulkRelationships(ctx context.Context, relationships []RelationshipObject) error
}

// Repository defines the interface for entity CRUD operations.
// Only used when code is generated with --with-repository.
type Repository interface {
	Create(ctx context.Context, entityType Type, id ID, data any) error
	Get(ctx context.Context, entityType Type, id ID) (any, error)
	Update(ctx context.Context, entityType Type, id ID, data any) error
	Delete(ctx context.Context, entityType Type, id ID) error
	Exists(ctx context.Context, entityType Type, id ID) (bool, error)
	List(ctx context.Context, entityType Type, filters map[string]any) ([]ID, error)
}
