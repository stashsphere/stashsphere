{ stdenv
, nodejs
, pnpm
, version
, writeText
, apiHost ? "https://api.stashsphere.com"
}:
let
  config = writeText "config.json" (builtins.toJSON ({
    inherit apiHost;
  }));
in
stdenv.mkDerivation (finalAttrs: {
  pname = "stashsphere";
  inherit version;

  src = ./.;

  nativeBuildInputs = [
    nodejs
    pnpm.configHook
  ];

  pnpmDepsHash = "sha256-XLvbQd7UpjLqono6+1KQQsRQUfjhi2kiA7AvZ8f8eO8=";

  pnpmDeps = pnpm.fetchDeps {
    inherit (finalAttrs) pname version src;
    hash = finalAttrs.pnpmDepsHash;
  };

  buildPhase = ''
    runHook preBuild
    
    pnpm build
    
    runHook postBuild
  '';

  installPhase = ''
    mkdir -p $out
    cp -r dist $out/.
    if [[ "${apiHost}" != "" ]]
    then
      cp ${config} $out/dist/config.json
    fi
  '';
})
