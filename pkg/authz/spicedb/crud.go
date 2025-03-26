package spicedb

import (
	"github.com/danhtran94/authzed-codegen/pkg/authz"
)

type Engine struct{}

var _ authz.Engine = &Engine{}

func (e *Engine) CreateRelations(to authz.Resource, relation authz.Relation, subject authz.Type, ids []authz.ID) error {
	return nil
}

func (e *Engine) DeleteRelations(from authz.Resource, relation authz.Relation, subject authz.Type, ids []authz.ID) error {
	return nil
}

func (e *Engine) CheckPermission(dest authz.Resource, has authz.Permission, subject authz.Type, audIDs []authz.ID) error {
	return nil
}

func (e *Engine) LookupResources(from authz.Type, match authz.Permission, subject authz.Type, byIDs []authz.ID) ([]authz.ID, error) {
	return []authz.ID{}, nil
}

func (e *Engine) LookupSubjects(on authz.Resource, permission authz.Permission, subject authz.Type) ([]authz.ID, error) {
	return []authz.ID{}, nil
}

func (e *Engine) ReadRelations(from authz.Resource, relation authz.Relation, subject authz.Type) ([]authz.ID, error) {
	return []authz.ID{}, nil
}
