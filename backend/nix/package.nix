{ buildGoModule
, version
, file
}:
buildGoModule {
  pname = "stashsphere-backend";
  inherit version;

  src = builtins.filterSource (path: type: baseNameOf path != "nix") ../.;

  vendorHash = "sha256-9+cXpqDZsxHPKR9TYi6h7JkgARlLVQ9n7rRQhllLzsg=";

  buildInputs = [
    # libmagic
    file
  ];
  
  doCheck = false;
}
