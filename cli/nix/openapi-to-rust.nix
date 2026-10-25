{
  lib,
  rustPlatform,
  fetchFromGitHub,
  versionCheckHook,
}:

# not yet packaged in nixpkgs
rustPlatform.buildRustPackage (finalAttrs: {
  pname = "openapi-to-rust";
  version = "0.23.0";

  src = fetchFromGitHub {
    owner = "gpu-cli";
    repo = "openapi-to-rust";
    tag = "v${finalAttrs.version}";
    hash = "sha256-tVNgRJ43meSdqOXlGBqvjY0g/7tdxLVsSlESXG/ayeo=";
  };

  cargoHash = "sha256-yhX4xn/QTeMHJ7kq65Qmhts8VeZkJBmaONkqp80LJU8=";

  cargoBuildFlags = [
    "--bin"
    "openapi-to-rust"
  ];

  # many tests compile the generated code with cargo, which needs crates.io
  doCheck = false;

  nativeInstallCheckInputs = [ versionCheckHook ];
  doInstallCheck = true;

  meta = {
    description = "Generate typed Rust models, HTTP clients and Axum servers from OpenAPI 3.0/3.1";
    homepage = "https://github.com/gpu-cli/openapi-to-rust";
    changelog = "https://github.com/gpu-cli/openapi-to-rust/blob/v${finalAttrs.version}/CHANGELOG.md";
    license = lib.licenses.mit;
    mainProgram = "openapi-to-rust";
  };
})
