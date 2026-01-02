final: prev: {
  stashsphereFrontendVersion = "0.1";
  stashsphereFrontend = with final; final.callPackage ./default.nix { version=stashsphereFrontendVersion; };
}
