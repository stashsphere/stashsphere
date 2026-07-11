{
  buildGoModule,
  version,
  file,
  postgresql,
  postgresqlTestHook,
}:
buildGoModule {
  pname = "stashsphere-backend";
  inherit version;

  src = builtins.path {
    name = "backend-src";
    filter = (path: type: baseNameOf path != "nix");
    path = ../.;
  };

  vendorHash = "sha256-OOrMtJEW8lNm89/VEmNNKJzbiQhPIhNvGa0U3p6MMc0=";

  buildInputs = [
    # libmagic
    file
  ];

  doCheck = true;

  nativeCheckInputs = [
    postgresql
    postgresqlTestHook
  ];

  outputs = [
    "out"
    "doc"
  ];

  postInstall = ''
    mkdir -p $doc
    $out/bin/backend openapi-dump --output $doc/openapi.json
  '';
}
