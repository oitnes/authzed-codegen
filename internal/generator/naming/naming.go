package naming

import (
	"strings"
	"unicode"
)

// ToPascalCase converts a name to PascalCase.
// Handles namespace separators (/) and underscores.
// e.g., "bookingsvc/booking" -> "BookingsvcBooking", "public_forum" -> "PublicForum"
func ToPascalCase(s string) string {
	if strings.Contains(s, "/") {
		parts := strings.Split(s, "/")
		for i, part := range parts {
			parts[i] = ToPascalCase(part)
		}
		return strings.Join(parts, "")
	}

	parts := strings.Split(s, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		result.WriteString(string(runes))
	}
	return result.String()
}

// ToCamelCase converts a name to camelCase.
// e.g., "public_forum" -> "publicForum"
func ToCamelCase(s string) string {
	pascal := ToPascalCase(s)
	if len(pascal) == 0 {
		return pascal
	}
	runes := []rune(pascal)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// ToSnakeCase converts a name to snake_case suitable for filenames.
// Replaces "/" with "_" and converts to lowercase.
// e.g., "bookingsvc/booking" -> "bookingsvc_booking"
func ToSnakeCase(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "/", "_"))
}

// TypeConstName generates the constant name for a type.
// e.g., "bookingsvc/booking" -> "TypeBookingsvcBooking"
func TypeConstName(typeName string) string {
	return "Type" + ToPascalCase(typeName)
}

// RelationConstName generates the constant name for a relation.
// e.g., def="public_forum", rel="owner" -> "PublicForumRelationOwner"
func RelationConstName(defName, relName string) string {
	return ToPascalCase(defName) + "Relation" + ToPascalCase(relName)
}

// PermissionConstName generates the constant name for a permission.
// e.g., def="public_forum", perm="view" -> "PublicForumPermissionView"
func PermissionConstName(defName, permName string) string {
	return ToPascalCase(defName) + "Permission" + ToPascalCase(permName)
}

// TypeStructName generates the struct type name for a definition.
// e.g., "public_forum" -> "PublicForum", "bookingsvc/booking" -> "BookingsvcBooking"
func TypeStructName(typeName string) string {
	return ToPascalCase(typeName)
}

// RelationObjectsStructName generates the input struct name for relation operations.
// e.g., def="public_forum", rel="owner" -> "PublicForumOwnerObjects"
func RelationObjectsStructName(defName, relName string) string {
	return ToPascalCase(defName) + ToPascalCase(relName) + "Objects"
}

// CheckInputStructName generates the input struct name for permission checks.
// e.g., def="public_forum", perm="view" -> "CheckPublicForumViewInputs"
func CheckInputStructName(defName, permName string) string {
	return "Check" + ToPascalCase(defName) + ToPascalCase(permName) + "Inputs"
}

// ReceiverName generates a short receiver variable name from a type name.
// Uses the lowercase initials of each PascalCase word, max 3 chars.
// e.g., "PublicForum" -> "pf", "BookingsvcBooking" -> "bb", "User" -> "u"
func ReceiverName(typeName string) string {
	var result strings.Builder
	pascal := ToPascalCase(typeName)
	for _, r := range pascal {
		if unicode.IsUpper(r) {
			result.WriteRune(unicode.ToLower(r))
		}
	}
	name := result.String()
	if len(name) == 0 {
		return "x"
	}
	if len(name) > 3 {
		return name[:3]
	}
	return name
}
