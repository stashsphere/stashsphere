{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };
  outputs = { self, nixpkgs }:
    let
      # from https://github.com/NixOS/templates/blob/master/go-hello/flake.nix
      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";

      version = builtins.substring 0 8 lastModifiedDate;

      supportedSystems = [ "x86_64-linux" "aarch64-linux" ];

      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      packages = forAllSystems
        (system:
          let
            pkgs = nixpkgsFor.${system};
          in
          {
            frontend = pkgs.callPackage ./default.nix { inherit version; };
          });
      overlay = import ./overlay.nix;
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};

          create-logo-and-favicon = pkgs.writeShellApplication {
            name = "create-logo-and-favicon";

            runtimeInputs = [ pkgs.imagemagick ];

            text = ''
              set -x
              magick -density 300 -define icon:auto-resize=256,128,96,64,48,32,16 -background none "$1" public/favicon.ico
              magick -background none -size 256x256 "$1" src/assets/stashsphere-logo-256.png
              cp "$1" public/icon.svg
            '';
          };
        in
        {
          # dev shell frontend
          default = pkgs.mkShell {
            name = "default";
            buildInputs = with pkgs; [
              create-logo-and-favicon
              nodePackages.npm
              nodejs_22
              pnpm
            ];
          };
        });
    };
}
