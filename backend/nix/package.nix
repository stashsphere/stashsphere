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

  vendorHash = "sha256-JtMQEcNFA7Z5QzBOrUWH8w2j/2yr4F1FTmdos+U4M6I=";

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
