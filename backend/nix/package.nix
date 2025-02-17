{ buildGoModule
, version
, file
}:
buildGoModule {
  pname = "stashsphere-backend";
  inherit version;

  src = builtins.filterSource (path: type: baseNameOf path != "nix") ../.;

  vendorHash = "sha256-pxWMQ65680mJv3qTgWwozKTrDIyP1/BrAXwlyCV9dpg=";

  buildInputs = [
    # libmagic
    file
  ];
  
  doCheck = false;
}
