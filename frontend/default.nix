{
  stdenv,
  nodejs,
  pnpm,
  version,
  writeText,
  apiHost ? "https://api.stashsphere.com",
  gitRevision ? "unknown",
}:
let
  config = writeText "config.json" (
    builtins.toJSON ({
      inherit apiHost;
    })
  );
  versionInfo = writeText "version.json" (
    builtins.toJSON ({
      inherit version gitRevision;
    })
  );
in
stdenv.mkDerivation (finalAttrs: {
  pname = "stashsphere";
  inherit version;

  src = ./.;

  nativeBuildInputs = [
    nodejs
    pnpm.configHook
  ];

  pnpmDepsHash = "sha256-6XBsZj84M5aqDvLkyZuUyWJ1k2GIU5vVqU/yiazBqio=";

  pnpmDeps = pnpm.fetchDeps {
    inherit (finalAttrs) pname version src;
    fetcherVersion = 3;
    hash = finalAttrs.pnpmDepsHash;
  };

  buildPhase = ''
    runHook preBuild

    cp ${versionInfo} src/version.json

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
