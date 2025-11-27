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

  vendorHash = "sha256-aV2Jd5wF5TTFtX8XdweZClFGFyUuNkKjeXep/8kWX7g=";

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
