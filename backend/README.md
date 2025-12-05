# StashSphere

## Config

Check `config/config.go` for the full set of options.
`stashsphere.yaml` contains a sample file for local development.
`.yaml` files are chainable so that secrets and general configs
can be separated.

## Development

### Building the application

You need `file` / `libmagic`.

```
go build -o stashsphere
```

### Running the application

You need a postgresql server running.

For example:

```
./stashsphere serve --conf stashsphere.yaml --conf invite.yaml
```

### OpenAPI

An OpenAPI 3.1 schema is generated through `fuego` dynamically from code.
For that, run the application with the `--serve-openapi` flag and then
navigate to `http://127.0.0.1:1323/swagger`.

Alternatively, use the `openapi-dump` command:

```
./stashsphere openapi-dump --output $(date +%s)-stashsphere.json
```

## Nix

### Build the application

`nix build .#packages.x86_64-linux.stashsphere`

This also run a `checkPhase`, i.e. `go test ./...`.

### Run the NixOS test

`nix build .#checks.x86_64-linux.nixos-test`

## License

AGPLv3

Copyright 2025 `Maximilian Güntner <code@mguentner.de>`