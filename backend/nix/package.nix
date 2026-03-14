{ buildGoModule
, version
, file
, postgresql
, postgresqlTestHook
}:
buildGoModule {
  pname = "stashsphere-backend";
  inherit version;

  src = builtins.filterSource (path: type: baseNameOf path != "nix") ../.;

  vendorHash = "sha256-g6Xwe9IBQW9AlvjdQ2gwxrMoce+EI5GeV9++6azjq24=";

  buildInputs = [
    # libmagic
    file
  ];
  
  doCheck = true;

  nativeCheckInputs = [
    postgresql
    postgresqlTestHook
  ];

  outputs = [ "out" "doc" ];

  postInstall = ''
    mkdir -p $doc
    $out/bin/backend openapi-dump --output $doc/openapi.json
  '';
}
