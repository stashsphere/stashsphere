final: prev: {
  # default value, this will be overwritten by the flake
  stashsphereVersion = "0.1";
  stashsphere = with final; final.callPackage ./package.nix { version=stashsphereVersion; };
  stashsphere-openapi = final.stashsphere.doc;
}
