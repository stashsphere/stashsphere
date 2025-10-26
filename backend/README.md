# StashSphere

## Config

Check `config/config.go` for the full set of options.
`stashsphere.yaml` contains a sample file for local development.
`.yaml` files are chainable so that secrets and general configs
can be separated.

## Development

### Building the application

`go build`. You need `file` / `libmagic`.

### Running the application

You need a postgresql server running.

For example:

`./backend serve --conf stashsphere.yaml --conf invite.yaml`

## Nix

### Build the application

`nix build .#packages.x86_64-linux.stashsphere`

This also run a `checkPhase`, i.e. `go test ./...`.

### Run the NixOS test

`nix build .#checks.x86_64-linux.nixos-test`

## License

AGPLv3

Copyright 2025 `Maximilian Güntner <code@mguentner.de>`