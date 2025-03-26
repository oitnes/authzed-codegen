# `authzed-codegen`

This repository contains Type-Safe Code Generation tools for AuthZed.

TLDR; `zed` schema => `go` code generation.

[Authzed](https://authzed.com/) is a powerful authorization engine that allows you to define and manage your authorization policies. This code generation tool helps you generate type-safe code for your AuthZed schemas, making it easier to work with your authorization policies in a type-safe manner.

## Installation

To install the dependencies, run:

```sh
go install -v github.com/danhtran94/authzed-codegen/cmd/authzed-codegen@latest
```

## Usage

To generate code, run:

```sh
authzed-codegen --output example/authzed example/schema.zed
```

## Features (check the `example` folder for more details)

- Create relationships and permissions stubs.
- Read relationships.
- Lookup permissions and relationships.
- Generate code for multiple languages (Go, ~~TypeScript, etc.~~).

## TODOs

- [ ] Support all schema features.
- [ ] Implement SpiceDB engine client.


## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.