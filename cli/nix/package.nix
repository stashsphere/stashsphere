{
  lib,
  rustPlatform,
  version,
}:

rustPlatform.buildRustPackage {
  pname = "stashctl";
  inherit version;

  src = lib.fileset.toSource {
    root = ../.;
    fileset = lib.fileset.unions [
      ../Cargo.toml
      ../Cargo.lock
      ../src
    ];
  };

  cargoLock.lockFile = ../Cargo.lock;

  meta = {
    description = "Command line client for StashSphere";
    license = lib.licenses.agpl3Only;
    mainProgram = "stashctl";
  };
}
