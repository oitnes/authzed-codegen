package spicedb

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/danhtran94/authzed-codegen/pkg/authz"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
)

type Engine struct {
	client *authzed.Client

	durationExpire time.Duration
	token          string
	setTokenTime   int64
}

var _ authz.Engine = &Engine{}

func NewEngine(client *authzed.Client, durationExpireToken time.Duration) *Engine {
	if durationExpireToken == 0 {
		durationExpireToken = 3 * time.Second
	}

	return &Engine{
		client:         client,
		durationExpire: durationExpireToken,
	}
}

func (e *Engine) setToken(token string) {
	e.token = token
	e.setTokenTime = time.Now().UnixNano()
}

func (e *Engine) getConsistencySnapshot() *v1.Consistency {
	now := time.Now().UnixNano()
	if now-e.setTokenTime < e.durationExpire.Nanoseconds() {
		return nil
	}

	return &v1.Consistency{
		Requirement: &v1.Consistency_AtExactSnapshot{
			AtExactSnapshot: &v1.ZedToken{
				Token: e.token,
			},
		},
	}
}

func (e *Engine) CreateRelations(ctx context.Context, to authz.Resource, relation authz.Relation, subject authz.Type, ids []authz.ID) error {
	updates := make([]*v1.RelationshipUpdate, 0, len(ids))
	for _, id := range ids {
		updates = append(updates, &v1.RelationshipUpdate{
			Operation: v1.RelationshipUpdate_OPERATION_CREATE,
			Relationship: &v1.Relationship{
				Resource: &v1.ObjectReference{
					ObjectType: string(to.Type),
					ObjectId:   string(to.ID),
				},
				Relation: string(relation),
				Subject: &v1.SubjectReference{
					Object: &v1.ObjectReference{
						ObjectType: string(subject),
						ObjectId:   string(id),
					},
				},
			},
		})
	}

	res, err := e.client.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{
		Updates: updates,
	})
	if err != nil {
		return err
	}

	e.setToken(res.WrittenAt.Token)

	return nil
}

func (e *Engine) DeleteRelations(ctx context.Context, from authz.Resource, relation authz.Relation, subject authz.Type, ids []authz.ID) error {
	updates := make([]*v1.RelationshipUpdate, 0, len(ids))
	for _, id := range ids {
		updates = append(updates, &v1.RelationshipUpdate{
			Operation: v1.RelationshipUpdate_OPERATION_DELETE,
			Relationship: &v1.Relationship{
				Resource: &v1.ObjectReference{
					ObjectType: string(from.Type),
					ObjectId:   string(from.ID),
				},
				Relation: string(relation),
				Subject: &v1.SubjectReference{
					Object: &v1.ObjectReference{
						ObjectType: string(subject),
						ObjectId:   string(id),
					},
				},
			},
		})
	}

	res, err := e.client.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{
		Updates: updates,
	})
	if err != nil {
		return err
	}

	e.setToken(res.WrittenAt.Token)

	return nil
}

func (e *Engine) CheckPermission(ctx context.Context, dest authz.Resource, has authz.Permission, subject authz.Type, audIDs []authz.ID) error {
	consistency := e.getConsistencySnapshot()

	for _, id := range audIDs {
		err := errorIfDenied(e.client.CheckPermission(ctx, &v1.CheckPermissionRequest{
			Consistency: consistency,
			Resource: &v1.ObjectReference{
				ObjectType: string(dest.Type),
				ObjectId:   string(dest.ID),
			},
			Permission: string(has),
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectType: string(subject),
					ObjectId:   string(id),
				},
			},
		}))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) LookupResources(ctx context.Context, from authz.Type, match authz.Permission, subject authz.Type, byIDs []authz.ID) ([]authz.ID, error) {
	consistency := e.getConsistencySnapshot()
	ids := []authz.ID{}

	for _, id := range byIDs {
		res, err := e.client.LookupResources(ctx, &v1.LookupResourcesRequest{
			Consistency:        consistency,
			ResourceObjectType: string(from),
			Permission:         string(match),
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectType: string(subject),
					ObjectId:   string(id),
				},
			},
		})

		if err != nil {
			return nil, err
		}

		data, err := res.Recv()
		for ; err == nil && data != nil; data, err = res.Recv() {
			ids = append(ids, authz.ID(data.ResourceObjectId))
		}

		if !errors.Is(err, io.EOF) {
			return nil, err
		}
	}

	return ids, nil
}

func (e *Engine) LookupSubjects(ctx context.Context, on authz.Resource, permission authz.Permission, subject authz.Type) ([]authz.ID, error) {
	consistency := e.getConsistencySnapshot()
	ids := []authz.ID{}

	res, err := e.client.LookupSubjects(ctx, &v1.LookupSubjectsRequest{
		Consistency: consistency,
		Resource: &v1.ObjectReference{
			ObjectType: string(on.Type),
			ObjectId:   string(on.ID),
		},
		Permission:        string(permission),
		SubjectObjectType: string(subject),
	})
	if err != nil {
		return nil, err
	}

	data, err := res.Recv()
	for ; err == nil && data != nil; data, err = res.Recv() {
		ids = append(ids, authz.ID(data.Subject.SubjectObjectId))
	}
	if !errors.Is(err, io.EOF) {
		return nil, err
	}

	return ids, nil
}

func (e *Engine) ReadRelations(ctx context.Context, from authz.Resource, relation authz.Relation, subject authz.Type) ([]authz.ID, error) {
	consistency := e.getConsistencySnapshot()
	ids := []authz.ID{}

	res, err := e.client.ReadRelationships(ctx, &v1.ReadRelationshipsRequest{
		Consistency: consistency,
		RelationshipFilter: &v1.RelationshipFilter{
			ResourceType:       string(from.Type),
			OptionalResourceId: string(from.ID),
			OptionalRelation:   string(relation),
			OptionalSubjectFilter: &v1.SubjectFilter{
				SubjectType: string(subject),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	data, err := res.Recv()
	for ; err == nil && data != nil; data, err = res.Recv() {
		ids = append(ids, authz.ID(data.Relationship.Subject.Object.ObjectId))
	}
	if !errors.Is(err, io.EOF) {
		return nil, err
	}

	return ids, nil
}

func errorIfDenied(res *v1.CheckPermissionResponse, err error) error {
	if err != nil {
		return err
	}

	if res.Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
		return nil
	}

	return authz.ErrPermissionDenied
}
