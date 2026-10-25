{
  lib,
  runCommand,
  diffutils,
  openapi-to-rust,
  stashsphere-openapi,
}:

let
  src = lib.fileset.toSource {
    root = ../.;
    fileset = lib.fileset.unions [
      ../openapi-to-rust.toml
      ../openapi.json
      ../src/api
    ];
  };
in
# fails when cli/openapi.json or the generated client in cli/src/api are
# out of date, see cli/openapi-to-rust.toml for how to regenerate them
runCommand "stashctl-codegen-check"
  {
    nativeBuildInputs = [
      diffutils
      openapi-to-rust
    ];
  }
  ''
    cp -r ${src} cli
    chmod -R u+w cli
    if ! diff -u cli/openapi.json ${stashsphere-openapi}/openapi.json; then
      echo "cli/openapi.json differs from the backend's OpenAPI spec" >&2
      exit 1
    fi
    openapi-to-rust generate --config cli/openapi-to-rust.toml --check
    touch $out
  ''
