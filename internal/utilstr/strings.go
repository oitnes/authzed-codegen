package utilstr

import (
	"fmt"
	"strings"
)

func UpperFirst(s string) string {
	if len(s) == 0 {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

func PackageName(s string) string {
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return strings.ToLower(s)
}

func Quote(s string) string {
	return "\"" + s + "\""
}

func SnakeToPascal(s string) string {
	var b strings.Builder
	words := strings.Split(s, "_")
	for _, word := range words {
		b.WriteString(UpperFirst(word))
	}
	return b.String()
}

func Length(list []any) string {
	return fmt.Sprintf("%d", len(list))
}

func TypeName(objectType string) string {
	name := strings.Split(objectType, "/")
	if len(name) == 0 {
		return ""
	}
	if len(name) == 1 {
		return UpperFirst(PackageName(name[0]))
	}
	return UpperFirst(PackageName(name[1]))
}
