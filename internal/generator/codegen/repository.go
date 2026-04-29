package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/oitnes/authzed-codegen/internal/generator/ast"
	"github.com/oitnes/authzed-codegen/internal/generator/naming"
)

// generateRepositoryMethods generates optional entity CRUD methods.
// Only called when Options.WithRepository is true.
func generateRepositoryMethods(f *jen.File, def *ast.Definition) {
	typeName := naming.TypeStructName(def.Name)
	receiver := naming.ReceiverName(typeName)
	typeConst := naming.TypeConstName(def.Name)

	// Create function (package-level)
	createFunc := "Create" + typeName
	f.Commentf("%s creates a new %s entity.", createFunc, def.Name)
	f.Func().Id(createFunc).Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("id").Id(typeName),
		jen.Id("data").Any(),
	).Error().Block(
		jen.Return(jen.Id("id").Dot("repo").Dot("Create").Call(
			jen.Id("ctx"),
			jen.Id(typeConst),
			jen.Qual(authzPkg, "ID").Call(jen.Id("id").Dot("id")),
			jen.Id("data"),
		)),
	)
	f.Line()

	// Get function (package-level)
	getFunc := "Get" + typeName
	f.Commentf("%s retrieves a %s entity by ID.", getFunc, def.Name)
	f.Func().Id(getFunc).Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("id").Id(typeName),
	).Params(jen.Any(), jen.Error()).Block(
		jen.Return(jen.Id("id").Dot("repo").Dot("Get").Call(
			jen.Id("ctx"),
			jen.Id(typeConst),
			jen.Qual(authzPkg, "ID").Call(jen.Id("id").Dot("id")),
		)),
	)
	f.Line()

	// Update method
	f.Commentf("Update updates this %s entity.", def.Name)
	f.Func().Params(jen.Id(receiver).Id(typeName)).Id("Update").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("data").Any(),
	).Error().Block(
		jen.Return(jen.Id(receiver).Dot("repo").Dot("Update").Call(
			jen.Id("ctx"),
			jen.Id(typeConst),
			jen.Qual(authzPkg, "ID").Call(jen.Id(receiver).Dot("id")),
			jen.Id("data"),
		)),
	)
	f.Line()

	// Delete method
	f.Commentf("Delete removes this %s entity.", def.Name)
	f.Func().Params(jen.Id(receiver).Id(typeName)).Id("Delete").Params(
		jen.Id("ctx").Qual("context", "Context"),
	).Error().Block(
		jen.Return(jen.Id(receiver).Dot("repo").Dot("Delete").Call(
			jen.Id("ctx"),
			jen.Id(typeConst),
			jen.Qual(authzPkg, "ID").Call(jen.Id(receiver).Dot("id")),
		)),
	)
	f.Line()

	// Exists function (package-level)
	existsFunc := "Check" + typeName + "Exists"
	f.Commentf("%s checks if a %s entity exists.", existsFunc, def.Name)
	f.Func().Id(existsFunc).Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("id").Id(typeName),
	).Params(jen.Bool(), jen.Error()).Block(
		jen.Return(jen.Id("id").Dot("repo").Dot("Exists").Call(
			jen.Id("ctx"),
			jen.Id(typeConst),
			jen.Qual(authzPkg, "ID").Call(jen.Id("id").Dot("id")),
		)),
	)
	f.Line()

	// List function (package-level)
	listFunc := "List" + typeName + "s"
	f.Commentf("%s retrieves all %s entities matching the filters.", listFunc, def.Name)
	f.Func().Id(listFunc).Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("engine").Qual(authzPkg, "Engine"),
		jen.Id("repo").Qual(authzPkg, "Repository"),
		jen.Id("filters").Map(jen.String()).Any(),
	).Params(jen.Index().Id(typeName), jen.Error()).Block(
		jen.List(jen.Id("ids"), jen.Err()).Op(":=").Id("repo").Dot("List").Call(
			jen.Id("ctx"),
			jen.Id(typeConst),
			jen.Id("filters"),
		),
		jen.If(jen.Err().Op("!=").Nil()).Block(
			jen.Return(jen.Nil(), jen.Err()),
		),
		jen.Id("result").Op(":=").Make(jen.Index().Id(typeName), jen.Len(jen.Id("ids"))),
		jen.For(jen.Id("i").Op(",").Id("id").Op(":=").Range().Id("ids")).Block(
			jen.Id("result").Index(jen.Id("i")).Op("=").Id("New"+typeName).Call(jen.String().Call(jen.Id("id")), jen.Id("engine"), jen.Id("repo")),
		),
		jen.Return(jen.Id("result"), jen.Nil()),
	)
	f.Line()
}
