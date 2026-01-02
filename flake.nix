{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };
  outputs =
    { self, nixpkgs }:
    let
      # from https://github.com/NixOS/templates/blob/master/go-hello/flake.nix
      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";

      version = builtins.substring 0 8 lastModifiedDate;

      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
      ];

      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      nixpkgsFor = forAllSystems (
        system:
        import nixpkgs {
          inherit system;
          overlays = [
            self.overlay
            (final: prev: { stashsphereVersion = version; })
          ];
        }
      );
    in
    {
      overlay = import ./nix/overlay.nix;
      packages = forAllSystems (system: {
        inherit (nixpkgsFor.${system}) stashsphere;
        backend = (nixpkgsFor.${system}).stashsphere;
        frontend = (nixpkgsFor.${system}).stashsphere-frontend;
      });
      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};

          go-migrate-pg = pkgs.go-migrate.overrideAttrs (oldAttrs: {
            tags = [ "postgres" ];
          });

          sqlboiler-fixed = pkgs.sqlboiler.overrideAttrs (oldAttrs: {
            patches = [ ./nix/patches/0001-sqlboiler-manytomany.patch ];
          });

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
          backend = pkgs.mkShell {
            name = "default";
            buildInputs = with pkgs; [
              go
              gopls
              bashInteractive
              file
              sqlboiler-fixed
              go-migrate-pg
              postgresql
              nodePackages.npm
            ];
            env = {
              "PGHOST" = "127.0.0.1";
              "PGPORT" = "5432";
              "PGUSER" = "stashsphere";
              "PGPASSWORD" = "secret";
              "PGDATABASE" = "stashsphere";
            };
          };
          frontend = pkgs.mkShell {
            name = "default";
            buildInputs = with pkgs; [
              create-logo-and-favicon
              nodePackages.npm
              nodejs_22
              pnpm
            ];
          };
        }
      );

      checks = forAllSystems (system: {
        nixos-test = nixpkgsFor.${system}.callPackage ./backend/nix/nixos-test.nix { };
      });
    };
}
