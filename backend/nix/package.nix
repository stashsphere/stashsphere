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

  vendorHash = "sha256-spN5Ti0JF1gXE7AHo7BBCfPrDL6SpsR/J0l28aZkI2Y=";

  buildInputs = [
    # libmagic
    file
  ];
  
  doCheck = true;

  nativeCheckInputs = [
    postgresql
    postgresqlTestHook
  ];
}
