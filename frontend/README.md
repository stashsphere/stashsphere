# StashSphere Frontend

This is the frontend of stashsphere built using React, Typescript and Vite.

## Config

The application is configured using a simple config object. It configures
first and foremost the API base url of the backend.
In production mode the config is fetched relative to the frontend location,
in development mode the config is static, see `src/config/config.development.ts`.

## Nix

The frontend can be built using the provided nix flake. This also allows to
generate the config on-the-fly, see the `apiHost` argument.

Build:

```
nix build .#frontend
```

## License

AGPL v3
