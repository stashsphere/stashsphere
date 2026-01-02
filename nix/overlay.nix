final: prev: {
  # default value, this will be overwritten by the flake
  stashsphereVersion = "0.1";
  stashsphere = with final; final.callPackage ../backend/nix/package.nix { version=stashsphereVersion; };
  stashsphere-openapi = final.stashsphere.doc;
  stashsphere-frontend = with final; final.callPackage ../frontend/default.nix { version=stashsphereVersion; };
}
