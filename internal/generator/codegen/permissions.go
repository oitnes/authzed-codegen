package codegen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/oitnes/authzed-codegen/internal/generator/ast"
	"github.com/oitnes/authzed-codegen/internal/generator/naming"
)

// generatePermissionMethods generates Check and Lookup methods for each permission.
func generatePermissionMethods(f *jen.File, def *ast.Definition, generatedLookups map[string]bool, withRepository bool) {
	subjectTypes := collectSubjectTypes(def)

	for _, perm := range def.Permissions {
		generateCheckInputStruct(f, def, perm, subjectTypes)
		generateCheckMethod(f, def, perm, subjectTypes)
		generateLookupMethods(f, def, perm, subjectTypes, generatedLookups, withRepository)
	}
}

// collectSubjectTypes returns all unique subject type names used in relations of the definition.
func collectSubjectTypes(def *ast.Definition) []string {
	seen := make(map[string]bool)
	var types []string

	for _, rel := range def.Relations {
		for _, st := range rel.SubjectTypes {
			if !seen[st.TypeName] {
				seen[st.TypeName] = true
				types = append(types, st.TypeName)
			}
		}
	}

	return types
}

// generateCheckInputStruct generates the input struct for permission checks.
func generateCheckInputStruct(f *jen.File, def *ast.Definition, perm *ast.Permission, subjectTypes []string) {
	structName := naming.CheckInputStructName(def.Name, perm.Name)

	var fields []jen.Code
	for _, st := range subjectTypes {
		fieldName := naming.TypeStructName(st)
		fields = append(fields, jen.Id(fieldName).Index().Id(fieldName))
	}

	f.Commentf("%s holds subjects for %s permission checks.", structName, perm.Name)
	f.Type().Id(structName).Struct(fields...)
	f.Line()
}

// generateCheckMethod generates the Check{Permission} method.
func generateCheckMethod(f *jen.File, def *ast.Definition, perm *ast.Permission, subjectTypes []string) {
	typeName := naming.TypeStructName(def.Name)
	receiver := naming.ReceiverName(typeName)
	methodName := "Check" + naming.ToPascalCase(perm.Name)
	structName := naming.CheckInputStructName(def.Name, perm.Name)
	permConst := naming.PermissionConstName(def.Name, perm.Name)

	var body []jen.Code

	for _, st := range subjectTypes {
		fieldName := naming.TypeStructName(st)
		typeConst := naming.TypeConstName(st)

		body = append(body,
			jen.For(jen.Id("_").Op(",").Id("s").Op(":=").Range().Id("subjects").Dot(fieldName)).Block(
				jen.List(jen.Id("ok"), jen.Err()).Op(":=").Id(receiver).Dot("engine").Dot("CheckPermission").Call(
					jen.Id("ctx"),
					jen.Id(receiver).Dot("resource").Call(),
					jen.Id(permConst),
					jen.Id(typeConst),
					jen.Qual(authzPkg, "ID").Call(jen.Id("s").Dot("id")),
				),
				jen.If(jen.Err().Op("!=").Nil()).Block(
					jen.Return(jen.False(), jen.Err()),
				),
				jen.If(jen.Id("ok")).Block(
					jen.Return(jen.True(), jen.Nil()),
				),
			),
		)
	}

	body = append(body, jen.Return(jen.False(), jen.Nil()))

	f.Commentf("%s checks if any subject has %s permission on this %s.", methodName, perm.Name, def.Name)
	f.Func().Params(jen.Id(receiver).Id(typeName)).Id(methodName).Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("subjects").Id(structName),
	).Params(jen.Bool(), jen.Error()).Block(body...)
	f.Line()
}

// generateLookupMethods generates LookupResources and LookupSubjects methods.
func generateLookupMethods(f *jen.File, def *ast.Definition, perm *ast.Permission, subjectTypes []string, generatedLookups map[string]bool, withRepository bool) {
	typeName := naming.TypeStructName(def.Name)
	receiver := naming.ReceiverName(typeName)
	permConst := naming.PermissionConstName(def.Name, perm.Name)
	typeConst := naming.TypeConstName(def.Name)

	for _, st := range subjectTypes {
		subjectTypeName := naming.TypeStructName(st)
		subjectTypeConst := naming.TypeConstName(st)

		// LookupResources — package-level function
		lookupResKey := fmt.Sprintf("LookupResources_%s_%s_%s", def.Name, perm.Name, st)
		if !generatedLookups[lookupResKey] {
			generatedLookups[lookupResKey] = true

			funcName := fmt.Sprintf("Lookup%ssWith%sBy%s", typeName, naming.ToPascalCase(perm.Name), subjectTypeName)

			params := []jen.Code{
				jen.Id("ctx").Qual("context", "Context"),
				jen.Id("engine").Qual(authzPkg, "Engine"),
				jen.Id("subject").Id(subjectTypeName),
			}
			newResourceCall := jen.Id("New"+typeName).Call(
				jen.String().Call(jen.Id("id")),
				jen.Id("engine"),
			)
			if withRepository {
				params = append(params, jen.Id("repo").Qual(authzPkg, "Repository"))
				newResourceCall = jen.Id("New"+typeName).Call(
					jen.String().Call(jen.Id("id")),
					jen.Id("engine"),
					jen.Id("repo"),
				)
			}

			f.Commentf("%s finds all %s resources where the subject has %s permission.", funcName, def.Name, perm.Name)
			f.Func().Id(funcName).Params(params...).Params(jen.Index().Id(typeName), jen.Error()).Block(
				jen.List(jen.Id("ids"), jen.Err()).Op(":=").Id("engine").Dot("LookupResources").Call(
					jen.Id("ctx"),
					jen.Id(typeConst),
					jen.Id(permConst),
					jen.Id(subjectTypeConst),
					jen.Qual(authzPkg, "ID").Call(jen.Id("subject").Dot("id")),
				),
				jen.If(jen.Err().Op("!=").Nil()).Block(
					jen.Return(jen.Nil(), jen.Err()),
				),
				jen.Id("result").Op(":=").Make(jen.Index().Id(typeName), jen.Len(jen.Id("ids"))),
				jen.For(jen.Id("i").Op(",").Id("id").Op(":=").Range().Id("ids")).Block(
					jen.Id("result").Index(jen.Id("i")).Op("=").Add(newResourceCall),
				),
				jen.Return(jen.Id("result"), jen.Nil()),
			)
			f.Line()
		}

		// LookupSubjects — method on the resource
		lookupSubKey := fmt.Sprintf("LookupSubjects_%s_%s_%s", def.Name, perm.Name, st)
		if !generatedLookups[lookupSubKey] {
			generatedLookups[lookupSubKey] = true

			methodName := fmt.Sprintf("Lookup%ssWith%s", subjectTypeName, naming.ToPascalCase(perm.Name))
			newSubjectCall := newEntityCall(subjectTypeName, receiver, withRepository)

			f.Commentf("%s finds all %s subjects that have %s permission on this %s.", methodName, st, perm.Name, def.Name)
			f.Func().Params(jen.Id(receiver).Id(typeName)).Id(methodName).Params(
				jen.Id("ctx").Qual("context", "Context"),
			).Params(jen.Index().Id(subjectTypeName), jen.Error()).Block(
				jen.List(jen.Id("ids"), jen.Err()).Op(":=").Id(receiver).Dot("engine").Dot("LookupSubjects").Call(
					jen.Id("ctx"),
					jen.Id(receiver).Dot("resource").Call(),
					jen.Id(permConst),
					jen.Id(subjectTypeConst),
				),
				jen.If(jen.Err().Op("!=").Nil()).Block(
					jen.Return(jen.Nil(), jen.Err()),
				),
				jen.Id("result").Op(":=").Make(jen.Index().Id(subjectTypeName), jen.Len(jen.Id("ids"))),
				jen.For(jen.Id("i").Op(",").Id("id").Op(":=").Range().Id("ids")).Block(
					jen.Id("result").Index(jen.Id("i")).Op("=").Add(newSubjectCall),
				),
				jen.Return(jen.Id("result"), jen.Nil()),
			)
			f.Line()
		}
	}
}
