package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/oitnes/authzed-codegen/internal/generator/ast"
	"github.com/oitnes/authzed-codegen/internal/generator/naming"
)

// generateConstants writes type, relation, and permission constants.
func generateConstants(f *jen.File, def *ast.Definition) {
	typeName := naming.TypeStructName(def.Name)
	typeConst := naming.TypeConstName(def.Name)

	// Type constant
	f.Commentf("%s is the SpiceDB type constant for %s.", typeConst, def.Name)
	f.Const().Id(typeConst).Op("=").Qual(authzPkg, "Type").Call(jen.Lit(def.Name))
	f.Line()

	// Relation constants
	for _, rel := range def.Relations {
		constName := naming.RelationConstName(def.Name, rel.Name)
		f.Commentf("%s is the relation constant for %s.%s.", constName, typeName, rel.Name)
		f.Const().Id(constName).Op("=").Qual(authzPkg, "Relation").Call(jen.Lit(rel.Name))
	}
	if len(def.Relations) > 0 {
		f.Line()
	}

	// Permission constants
	for _, perm := range def.Permissions {
		constName := naming.PermissionConstName(def.Name, perm.Name)
		f.Commentf("%s is the permission constant for %s.%s.", constName, typeName, perm.Name)
		f.Const().Id(constName).Op("=").Qual(authzPkg, "Permission").Call(jen.Lit(perm.Name))
	}
	if len(def.Permissions) > 0 {
		f.Line()
	}
}

// generateTypeDefinition writes the struct, constructor, and accessor methods.
func generateTypeDefinition(f *jen.File, def *ast.Definition, withRepository bool) {
	typeName := naming.TypeStructName(def.Name)
	receiver := naming.ReceiverName(typeName)

	// Struct definition
	fields := []jen.Code{
		jen.Id("id").String(),
		jen.Id("engine").Qual(authzPkg, "Engine"),
	}
	if withRepository {
		fields = append(fields, jen.Id("repo").Qual(authzPkg, "Repository"))
	}

	f.Commentf("%s represents a %s resource.", typeName, def.Name)
	f.Type().Id(typeName).Struct(fields...)
	f.Line()

	// Constructor
	constructorName := "New" + typeName
	params := []jen.Code{
		jen.Id("id").String(),
		jen.Id("engine").Qual(authzPkg, "Engine"),
	}
	structFields := jen.Dict{
		jen.Id("id"):     jen.Id("id"),
		jen.Id("engine"): jen.Id("engine"),
	}
	if withRepository {
		params = append(params, jen.Id("repo").Qual(authzPkg, "Repository"))
		structFields[jen.Id("repo")] = jen.Id("repo")
	}

	f.Commentf("%s creates a new %s instance.", constructorName, typeName)
	f.Func().Id(constructorName).Params(params...).Id(typeName).Block(
		jen.Return(jen.Id(typeName).Values(structFields)),
	)
	f.Line()

	// ID() accessor
	f.Commentf("ID returns the resource identifier.")
	f.Func().Params(jen.Id(receiver).Id(typeName)).Id("ID").Params().String().Block(
		jen.Return(jen.Id(receiver).Dot("id")),
	)
	f.Line()

	// String() method
	f.Commentf("String implements fmt.Stringer.")
	f.Func().Params(jen.Id(receiver).Id(typeName)).Id("String").Params().String().Block(
		jen.Return(jen.Id(receiver).Dot("id")),
	)
	f.Line()

	// resource() helper (unexported)
	f.Func().Params(jen.Id(receiver).Id(typeName)).Id("resource").Params().Qual(authzPkg, "Resource").Block(
		jen.Return(jen.Qual(authzPkg, "Resource").Values(jen.Dict{
			jen.Id("Type"): jen.Id(naming.TypeConstName(def.Name)),
			jen.Id("ID"):   jen.Qual(authzPkg, "ID").Call(jen.Id(receiver).Dot("id")),
		})),
	)
	f.Line()
}
