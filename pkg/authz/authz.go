package authz

import (
	"context"
	"fmt"
)

var ErrNoInput = fmt.Errorf("no input")

type Engine interface {
	CreateRelations(to Resource, relation Relation, subject Type, ids []ID) error
	CheckPermission(dest Resource, has Permission, subject Type, audIDs []ID) error
	LookupResources(from Type, match Permission, subject Type, byIDs []ID) ([]ID, error)
	LookupSubjects(on Resource, permission Permission, subject Type) ([]ID, error)
	ReadRelations(from Resource, relation Relation, subject Type) ([]ID, error)
	DeleteRelations(from Resource, relation Relation, subject Type, ids []ID) error
}

var DefaultEngine Engine = nil

func SetDefaultEngine(engine Engine) {
	DefaultEngine = engine
}

func GetEngine(ctx context.Context) Engine {
	return DefaultEngine
}

type Type string

type ID string

type Permission string

type Relation string

type Subject struct {
	Type Type
	IDs  []ID
}

type Resource struct {
	Type Type
	ID   ID
}

func IDs[T ~string](ids []T) []ID {
	result := []ID{}

	for _, id := range ids {
		result = append(result, ID(id))
	}

	return result
}

func FromIDs[T ~string](ids []ID) []T {
	result := []T{}

	for _, id := range ids {
		result = append(result, T(id))
	}

	return result
}
