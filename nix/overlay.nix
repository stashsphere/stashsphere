final: prev: {
  # default value, this will be overwritten by the flake
  stashsphereVersion = "0.1";
  stashsphereGitRevision = "unknown";
  stashsphere = with final; final.callPackage ../backend/nix/package.nix { version=stashsphereVersion; };
  stashsphere-openapi = final.stashsphere.doc;
  stashsphere-openapi-lint = final.callPackage ../backend/nix/openapi-lint.nix { };
  openapi-to-rust = final.callPackage ../cli/nix/openapi-to-rust.nix { };
  stashctl = final.callPackage ../cli/nix/package.nix { version = final.stashsphereVersion; };
  stashsphere-frontend = with final; final.callPackage ../frontend/default.nix {
    version = stashsphereVersion;
    gitRevision = stashsphereGitRevision;
  };
}
