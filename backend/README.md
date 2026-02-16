# StashSphere Backend

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

When developing locally over plain HTTP, auth cookies will be rejected by the
browser because they carry the `Secure` flag by default. Disable it with:

```
./stashsphere serve --conf stashsphere.yaml --disable-secure-cookies
```

Alternatively, set the environment variable:

```
STASHSPHERE_DISABLE_SECURE_COOKIES=true ./stashsphere serve --conf stashsphere.yaml
```

Or add it to your config file:

```yaml
auth:
  disableSecureCookies: true
```

If you are using the Nix dev shell (`nix develop`), `STASHSPHERE_DISABLE_SECURE_COOKIES`
is already set automatically via `flake.nix`.

> **Warning:** Never disable secure cookies in production.

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