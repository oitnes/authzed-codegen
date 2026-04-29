package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/oitnes/authzed-codegen/internal/generator/ast"
	"github.com/oitnes/authzed-codegen/internal/generator/naming"
)

// generateRelationMethods generates input structs and Create/Read/Delete methods for each relation.
func generateRelationMethods(f *jen.File, def *ast.Definition, withRepository bool) {
	for _, rel := range def.Relations {
		generateRelationObjectsStruct(f, def, rel)
		generateCreateRelation(f, def, rel)
		generateReadRelation(f, def, rel, withRepository)
		generateDeleteRelation(f, def, rel)
	}
}

// generateRelationObjectsStruct generates the input struct for a relation's subject types.
func generateRelationObjectsStruct(f *jen.File, def *ast.Definition, rel *ast.Relation) {
	structName := naming.RelationObjectsStructName(def.Name, rel.Name)

	var fields []jen.Code
	for _, st := range rel.SubjectTypes {
		fieldName := naming.TypeStructName(st.TypeName)
		fields = append(fields, jen.Id(fieldName).Index().Id(fieldName))
		if st.IsWildcard {
			fields = append(fields, jen.Id(fieldName+"Wildcard").Bool())
		}
	}

	f.Commentf("%s holds subjects for %s relation operations.", structName, rel.Name)
	f.Type().Id(structName).Struct(fields...)
	f.Line()
}

// generateCreateRelation generates the Create{Relation}Relations method.
func generateCreateRelation(f *jen.File, def *ast.Definition, rel *ast.Relation) {
	typeName := naming.TypeStructName(def.Name)
	receiver := naming.ReceiverName(typeName)
	methodName := "Create" + naming.ToPascalCase(rel.Name) + "Relations"
	structName := naming.RelationObjectsStructName(def.Name, rel.Name)
	relConst := naming.RelationConstName(def.Name, rel.Name)

	var body []jen.Code
	for _, st := range rel.SubjectTypes {
		fieldName := naming.TypeStructName(st.TypeName)
		typeConst := naming.TypeConstName(st.TypeName)
		wildcardField := fieldName + "Wildcard"

		body = append(body,
			jen.If(jen.Len(jen.Id("subjects").Dot(fieldName)).Op(">").Lit(0)).Block(
				jen.Id("ids").Op(":=").Make(jen.Index().Qual(authzPkg, "ID"), jen.Len(jen.Id("subjects").Dot(fieldName))),
				jen.For(jen.Id("i").Op(",").Id("s").Op(":=").Range().Id("subjects").Dot(fieldName)).Block(
					jen.Id("ids").Index(jen.Id("i")).Op("=").Qual(authzPkg, "ID").Call(jen.Id("s").Dot("id")),
				),
				jen.If(
					jen.Err().Op(":=").Id(receiver).Dot("engine").Dot("CreateRelations").Call(
						jen.Id("ctx"),
						jen.Id(receiver).Dot("resource").Call(),
						jen.Id(relConst),
						jen.Id(typeConst),
						jen.Id("ids"),
					),
					jen.Err().Op("!=").Nil(),
				).Block(
					jen.Return(jen.Err()),
				),
			),
		)

		if st.IsWildcard {
			body = append(body,
				jen.If(jen.Id("subjects").Dot(wildcardField)).Block(
					jen.If(
						jen.Err().Op(":=").Id(receiver).Dot("engine").Dot("CreateRelations").Call(
							jen.Id("ctx"),
							jen.Id(receiver).Dot("resource").Call(),
							jen.Id(relConst),
							jen.Id(typeConst),
							jen.Index().Qual(authzPkg, "ID").Values(jen.Qual(authzPkg, "ID").Call(jen.Lit("*"))),
						),
						jen.Err().Op("!=").Nil(),
					).Block(
						jen.Return(jen.Err()),
					),
				),
			)
		}
	}

	body = append(body, jen.Return(jen.Nil()))

	f.Commentf("%s creates %s relations for this %s.", methodName, rel.Name, def.Name)
	f.Func().Params(jen.Id(receiver).Id(typeName)).Id(methodName).Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("subjects").Id(structName),
	).Error().Block(body...)
	f.Line()
}

// generateReadRelation generates the Read{Relation}Relations method.
func generateReadRelation(f *jen.File, def *ast.Definition, rel *ast.Relation, withRepository bool) {
	typeName := naming.TypeStructName(def.Name)
	receiver := naming.ReceiverName(typeName)
	methodName := "Read" + naming.ToPascalCase(rel.Name) + "Relations"
	structName := naming.RelationObjectsStructName(def.Name, rel.Name)
	relConst := naming.RelationConstName(def.Name, rel.Name)

	var body []jen.Code
	body = append(body, jen.Var().Id("result").Id(structName))

	for _, st := range rel.SubjectTypes {
		fieldName := naming.TypeStructName(st.TypeName)
		typeConst := naming.TypeConstName(st.TypeName)
		idsVar := "ids" + fieldName
		wildcardField := fieldName + "Wildcard"

		newSubjectCall := jen.Id("New"+fieldName).Call(
			jen.String().Call(jen.Id("id")),
			jen.Id(receiver).Dot("engine"),
		)
		if withRepository {
			newSubjectCall = jen.Id("New"+fieldName).Call(
				jen.String().Call(jen.Id("id")),
				jen.Id(receiver).Dot("engine"),
				jen.Id(receiver).Dot("repo"),
			)
		}

		loopBody := []jen.Code{
			jen.Id("result").Dot(fieldName).Op("=").Append(
				jen.Id("result").Dot(fieldName),
				newSubjectCall,
			),
		}
		if st.IsWildcard {
			loopBody = []jen.Code{
				jen.If(jen.Id("id").Op("==").Qual(authzPkg, "ID").Call(jen.Lit("*"))).Block(
					jen.Id("result").Dot(wildcardField).Op("=").True(),
				).Else().Block(
					jen.Id("result").Dot(fieldName).Op("=").Append(
						jen.Id("result").Dot(fieldName),
						newSubjectCall,
					),
				),
			}
		}

		body = append(body,
			jen.List(jen.Id(idsVar), jen.Err()).Op(":=").Id(receiver).Dot("engine").Dot("ReadRelations").Call(
				jen.Id("ctx"),
				jen.Id(receiver).Dot("resource").Call(),
				jen.Id(relConst),
				jen.Id(typeConst),
			),
			jen.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(jen.Id(structName).Values(), jen.Err()),
			),
			jen.For(jen.Id("_").Op(",").Id("id").Op(":=").Range().Id(idsVar)).Block(loopBody...),
		)
	}

	body = append(body, jen.Return(jen.Id("result"), jen.Nil()))

	f.Commentf("%s reads %s relations for this %s.", methodName, rel.Name, def.Name)
	f.Func().Params(jen.Id(receiver).Id(typeName)).Id(methodName).Params(
		jen.Id("ctx").Qual("context", "Context"),
	).Params(jen.Id(structName), jen.Error()).Block(body...)
	f.Line()
}

// generateDeleteRelation generates the Delete{Relation}Relations method.
func generateDeleteRelation(f *jen.File, def *ast.Definition, rel *ast.Relation) {
	typeName := naming.TypeStructName(def.Name)
	receiver := naming.ReceiverName(typeName)
	methodName := "Delete" + naming.ToPascalCase(rel.Name) + "Relations"
	structName := naming.RelationObjectsStructName(def.Name, rel.Name)
	relConst := naming.RelationConstName(def.Name, rel.Name)

	var body []jen.Code
	for _, st := range rel.SubjectTypes {
		fieldName := naming.TypeStructName(st.TypeName)
		typeConst := naming.TypeConstName(st.TypeName)
		wildcardField := fieldName + "Wildcard"

		body = append(body,
			jen.If(jen.Len(jen.Id("subjects").Dot(fieldName)).Op(">").Lit(0)).Block(
				jen.Id("ids").Op(":=").Make(jen.Index().Qual(authzPkg, "ID"), jen.Len(jen.Id("subjects").Dot(fieldName))),
				jen.For(jen.Id("i").Op(",").Id("s").Op(":=").Range().Id("subjects").Dot(fieldName)).Block(
					jen.Id("ids").Index(jen.Id("i")).Op("=").Qual(authzPkg, "ID").Call(jen.Id("s").Dot("id")),
				),
				jen.If(
					jen.Err().Op(":=").Id(receiver).Dot("engine").Dot("DeleteRelations").Call(
						jen.Id("ctx"),
						jen.Id(receiver).Dot("resource").Call(),
						jen.Id(relConst),
						jen.Id(typeConst),
						jen.Id("ids"),
					),
					jen.Err().Op("!=").Nil(),
				).Block(
					jen.Return(jen.Err()),
				),
			),
		)

		if st.IsWildcard {
			body = append(body,
				jen.If(jen.Id("subjects").Dot(wildcardField)).Block(
					jen.If(
						jen.Err().Op(":=").Id(receiver).Dot("engine").Dot("DeleteRelations").Call(
							jen.Id("ctx"),
							jen.Id(receiver).Dot("resource").Call(),
							jen.Id(relConst),
							jen.Id(typeConst),
							jen.Index().Qual(authzPkg, "ID").Values(jen.Qual(authzPkg, "ID").Call(jen.Lit("*"))),
						),
						jen.Err().Op("!=").Nil(),
					).Block(
						jen.Return(jen.Err()),
					),
				),
			)
		}
	}

	body = append(body, jen.Return(jen.Nil()))

	f.Commentf("%s deletes %s relations for this %s.", methodName, rel.Name, def.Name)
	f.Func().Params(jen.Id(receiver).Id(typeName)).Id(methodName).Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("subjects").Id(structName),
	).Error().Block(body...)
	f.Line()
}
