package spicedb

import (
	"context"
	"io"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/oitnes/authzed-codegen/pkg/authz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Engine implements authz.Engine using a SpiceDB client.
type Engine struct {
	client *authzed.Client
}

// NewEngine creates a SpiceDB-backed Engine by connecting to the given endpoint
// with bearer-token auth. Additional gRPC dial options can be provided to
// override TLS settings or other transport configuration.
func NewEngine(endpoint, token string, opts ...grpc.DialOption) (*Engine, error) {
	defaults := []grpc.DialOption{
		grpcutil.WithInsecureBearerToken(token),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	client, err := authzed.NewClient(endpoint, append(defaults, opts...)...)
	if err != nil {
		return nil, err
	}
	return &Engine{client: client}, nil
}

// NewEngineWithClient creates a SpiceDB-backed Engine from an existing authzed
// client. Useful when you need custom TLS, retry policies, or in tests.
func NewEngineWithClient(client *authzed.Client) *Engine {
	return &Engine{client: client}
}

// ReadSchema returns the current schema text, or an empty string if no schema
// has been written yet.
func (e *Engine) ReadSchema(ctx context.Context) (string, error) {
	resp, err := e.client.ReadSchema(ctx, &v1.ReadSchemaRequest{})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return "", nil
		}
		return "", err
	}
	return resp.SchemaText, nil
}

// WriteSchema overwrites the SpiceDB schema unconditionally.
func (e *Engine) WriteSchema(ctx context.Context, schema string) error {
	_, err := e.client.WriteSchema(ctx, &v1.WriteSchemaRequest{Schema: schema})
	return err
}

// EnsureSchema writes the schema only when no schema exists yet. Use this
// during application startup to avoid overwriting a schema that was already
// applied by a previous deployment.
func (e *Engine) EnsureSchema(ctx context.Context, schema string) error {
	existing, err := e.ReadSchema(ctx)
	if err != nil {
		return err
	}
	if existing != "" {
		return nil
	}
	return e.WriteSchema(ctx, schema)
}

func (e *Engine) CreateRelations(ctx context.Context, resource authz.Resource, relation authz.Relation, subjectType authz.Type, subjectIDs []authz.ID) error {
	updates := make([]*v1.RelationshipUpdate, len(subjectIDs))
	for i, id := range subjectIDs {
		updates[i] = &v1.RelationshipUpdate{
			Operation: v1.RelationshipUpdate_OPERATION_CREATE,
			Relationship: &v1.Relationship{
				Resource: &v1.ObjectReference{
					ObjectType: string(resource.Type),
					ObjectId:   string(resource.ID),
				},
				Relation: string(relation),
				Subject: &v1.SubjectReference{
					Object: &v1.ObjectReference{
						ObjectType: string(subjectType),
						ObjectId:   string(id),
					},
				},
			},
		}
	}

	_, err := e.client.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{
		Updates: updates,
	})
	return err
}

func (e *Engine) ReadRelations(ctx context.Context, resource authz.Resource, relation authz.Relation, subjectType authz.Type) ([]authz.ID, error) {
	stream, err := e.client.ReadRelationships(ctx, &v1.ReadRelationshipsRequest{
		RelationshipFilter: &v1.RelationshipFilter{
			ResourceType:       string(resource.Type),
			OptionalResourceId: string(resource.ID),
			OptionalRelation:   string(relation),
			OptionalSubjectFilter: &v1.SubjectFilter{
				SubjectType: string(subjectType),
			},
		},
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true},
		},
	})
	if err != nil {
		return nil, err
	}

	var ids []authz.ID
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		ids = append(ids, authz.ID(resp.Relationship.Subject.Object.ObjectId))
	}

	return ids, nil
}

func (e *Engine) DeleteRelations(ctx context.Context, resource authz.Resource, relation authz.Relation, subjectType authz.Type, subjectIDs []authz.ID) error {
	updates := make([]*v1.RelationshipUpdate, len(subjectIDs))
	for i, id := range subjectIDs {
		updates[i] = &v1.RelationshipUpdate{
			Operation: v1.RelationshipUpdate_OPERATION_DELETE,
			Relationship: &v1.Relationship{
				Resource: &v1.ObjectReference{
					ObjectType: string(resource.Type),
					ObjectId:   string(resource.ID),
				},
				Relation: string(relation),
				Subject: &v1.SubjectReference{
					Object: &v1.ObjectReference{
						ObjectType: string(subjectType),
						ObjectId:   string(id),
					},
				},
			},
		}
	}

	_, err := e.client.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{
		Updates: updates,
	})
	return err
}

func (e *Engine) CheckPermission(ctx context.Context, resource authz.Resource, permission authz.Permission, subjectType authz.Type, subjectID authz.ID) (bool, error) {
	resp, err := e.client.CheckPermission(ctx, &v1.CheckPermissionRequest{
		Resource: &v1.ObjectReference{
			ObjectType: string(resource.Type),
			ObjectId:   string(resource.ID),
		},
		Permission: string(permission),
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: string(subjectType),
				ObjectId:   string(subjectID),
			},
		},
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true},
		},
	})
	if err != nil {
		return false, err
	}

	return resp.Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, nil
}

func (e *Engine) LookupResources(ctx context.Context, resourceType authz.Type, permission authz.Permission, subjectType authz.Type, subjectID authz.ID) ([]authz.ID, error) {
	stream, err := e.client.LookupResources(ctx, &v1.LookupResourcesRequest{
		ResourceObjectType: string(resourceType),
		Permission:         string(permission),
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: string(subjectType),
				ObjectId:   string(subjectID),
			},
		},
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true},
		},
	})
	if err != nil {
		return nil, err
	}

	var ids []authz.ID
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		ids = append(ids, authz.ID(resp.ResourceObjectId))
	}

	return ids, nil
}

func (e *Engine) LookupSubjects(ctx context.Context, resource authz.Resource, permission authz.Permission, subjectType authz.Type) ([]authz.ID, error) {
	stream, err := e.client.LookupSubjects(ctx, &v1.LookupSubjectsRequest{
		Resource: &v1.ObjectReference{
			ObjectType: string(resource.Type),
			ObjectId:   string(resource.ID),
		},
		Permission:        string(permission),
		SubjectObjectType: string(subjectType),
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true},
		},
	})
	if err != nil {
		return nil, err
	}

	var ids []authz.ID
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		ids = append(ids, authz.ID(resp.Subject.SubjectObjectId))
	}

	return ids, nil
}

func (e *Engine) CheckBulkPermission(ctx context.Context, checks []authz.PermissionCheck) ([]bool, error) {
	items := make([]*v1.CheckBulkPermissionsRequestItem, len(checks))
	for i, check := range checks {
		items[i] = &v1.CheckBulkPermissionsRequestItem{
			Resource: &v1.ObjectReference{
				ObjectType: string(check.Resource.Type),
				ObjectId:   string(check.Resource.ID),
			},
			Permission: string(check.Permission),
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectType: string(check.SubjectType),
					ObjectId:   string(check.SubjectID),
				},
			},
		}
	}

	resp, err := e.client.CheckBulkPermissions(ctx, &v1.CheckBulkPermissionsRequest{
		Items: items,
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true},
		},
	})
	if err != nil {
		return nil, err
	}

	results := make([]bool, len(resp.Pairs))
	for i, pair := range resp.Pairs {
		if pair.GetItem() != nil {
			results[i] = pair.GetItem().Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION
		}
	}

	return results, nil
}

func (e *Engine) ExportBulkRelationships(ctx context.Context, filter authz.RelationshipFilter) ([]authz.RelationshipObject, error) {
	stream, err := e.client.ExportBulkRelationships(ctx, &v1.ExportBulkRelationshipsRequest{
		OptionalRelationshipFilter: &v1.RelationshipFilter{
			ResourceType:       filter.ResourceType,
			OptionalResourceId: filter.ResourceID,
			OptionalRelation:   filter.Relation,
			OptionalSubjectFilter: &v1.SubjectFilter{
				SubjectType: filter.SubjectType,
			},
		},
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true},
		},
	})
	if err != nil {
		return nil, err
	}

	var relationships []authz.RelationshipObject
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, rel := range resp.Relationships {
			relationships = append(relationships, authz.RelationshipObject{
				Resource: authz.Resource{
					Type: authz.Type(rel.Resource.ObjectType),
					ID:   authz.ID(rel.Resource.ObjectId),
				},
				Relation:    authz.Relation(rel.Relation),
				SubjectType: authz.Type(rel.Subject.Object.ObjectType),
				SubjectID:   authz.ID(rel.Subject.Object.ObjectId),
			})
		}
	}

	return relationships, nil
}

func (e *Engine) ImportBulkRelationships(ctx context.Context, relationships []authz.RelationshipObject) error {
	stream, err := e.client.ImportBulkRelationships(ctx)
	if err != nil {
		return err
	}

	for _, rel := range relationships {
		req := &v1.ImportBulkRelationshipsRequest{
			Relationships: []*v1.Relationship{
				{
					Resource: &v1.ObjectReference{
						ObjectType: string(rel.Resource.Type),
						ObjectId:   string(rel.Resource.ID),
					},
					Relation: string(rel.Relation),
					Subject: &v1.SubjectReference{
						Object: &v1.ObjectReference{
							ObjectType: string(rel.SubjectType),
							ObjectId:   string(rel.SubjectID),
						},
					},
				},
			},
		}
		if err := stream.Send(req); err != nil {
			return err
		}
	}

	_, err = stream.CloseAndRecv()
	return err
}
