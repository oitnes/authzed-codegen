# `authzed-codegen`

This repository contains a Type-Safe Code API Call Generation tool for [SpiceDB](https://authzed.com/docs/spicedb) schemas.

[AuthZed](https://authzed.com/) is a powerful authorization engine built around [SpiceDB](https://spicedb.io/) that allows you to define and manage your authorization policies using the Zanzibar-inspired schema language. This code generation tool parses SpiceDB `.zed` schema files and generates type-safe Go code, making it easier to work with your authorization policies in a compile-time safe manner.

**TLDR**: `.zed` schema files → type-safe Go code generation.

## Installation

```bash
go install github.com/oitnes/authzed-codegen/cmd/authzed-codegen@latest
```

## Usage

### Command Line Options

- `--schema` or `-schema`: Path to the input `.zed` schema file (required)
- `--output` or `-output`: Output directory for generated Go files (required)
- `--package` or `-package`: Package name for generated code (optional; defaults to output directory name)
- `--with-repository` or `-with-repository`: Generate optional entity repository CRUD methods

### Example

Generate Go code from SpiceDB schema files:

```sh
authzed-codegen --schema path/to/schema.zed --output path/to/output/directory
```

Generate code with a custom package name:

```sh
authzed-codegen --schema path/to/schema.zed --output path/to/output/directory --package permissions
```

Generate code with repository CRUD helpers:

```sh
authzed-codegen --schema path/to/schema.zed --output path/to/output/directory --with-repository
```

## Features

### ✅ Supported SpiceDB Schema Features

- **Definitions**: Object type definitions with relations and permissions
- **Relations**: Type-safe relation definitions with union types (e.g., `user | customer`)
- **Permissions**: Complex permission expressions with full operator support:
  - `+` (Union): Combines relations/permissions
  - `-` (Exclusion): Removes subjects from a set
  - `&` (Intersection): Finds common subjects
  - `->` (Arrow): Traverses object hierarchies for nested permissions
  - `()` (Grouping): Parentheses for expression precedence
  - `|` (Union): Union of relation types
  - `:*` (Wildcard): Universal access patterns
- **Namespaces**: Support for prefixed definitions (e.g., `menusvc/order`, `bookingsvc/booking`) and regular (e.g., `order`)
- **Comments**: Line comments (`//`) and block comments (`/* */`)

### Generated Go Code Includes

- **Type-safe constants** for all object types, relations, and permissions
- **Struct types** for relationship objects and permission input validation
- **CRUD operations** for relationships:
  - `Create{Relation}Relations()` - Create new relationships
  - `Delete{Relation}Relations()` - Remove relationships
  - `Read{Relation}Relations()` - Read existing relationships, returning a struct with typed subject slices; wildcard relations also include a `{SubjectType}Wildcard bool` field
- **Permission checking** methods:
  - `Check{Permission}()` - Verify if permission is granted (method on resource type)
  - `Lookup{Type}sWith{Permission}By{SubjectType}()` - Find all resources of a type where a subject has a given permission (package-level function)
  - `Lookup{SubjectType}sWith{Permission}()` - Find all subjects that have a given permission on this resource (method on resource type)
- **Repository CRUD helpers** (generated with `--with-repository`):
  - `Create{Type}()` - Create a new entity (package-level)
  - `Get{Type}()` - Retrieve an entity by ID (package-level)
  - `Update()` - Update this entity (method)
  - `Delete()` - Delete this entity (method)
  - `Exists()` - Check if this entity exists (method)
  - `List{Type}s()` - List entities with optional filters (package-level)
- **Utility functions** for type conversion and ID management

## Dependencies

The generated code depends on the `authz` package which provides:
- SpiceDB client integration
- Type definitions for resources, relations, and permissions
- Runtime methods for authorization operations
