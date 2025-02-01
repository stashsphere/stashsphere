{ buildGoModule
, version
, file
}:
buildGoModule {
  pname = "stashsphere-backend";
  inherit version;

  src = ../.;

  vendorHash = "sha256-ij+xvCsVY94RPcjuW6hpVWWFWVfxKWBHa+diaesCLtg=";
  buildInputs = [
    # libmagic
    file
  ];
  
  doCheck = false;
}
