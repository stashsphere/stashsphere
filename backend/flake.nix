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

      nixpkgsFor = forAllSystems (system: import nixpkgs {
        inherit system;
        overlays = [
          self.overlay
          (final: prev: {stashsphereVersion = version; })
        ];
      });
    in
    {
      overlay = import ./nix/overlay.nix;
      packages = forAllSystems
        (system:
          {
            inherit (nixpkgsFor.${system}) stashsphere;
          });
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          # combined dev shell for backend and frontend
          default = pkgs.mkShell {
            name = "default";
            buildInputs = with pkgs; [
              go
              gopls
              bashInteractive
              file
              sqlboiler
              go-migrate
              postgresql
            ];
            env = {
              "DATABASE_URL" = "postgresql://stashsphere:secret@127.0.0.1:5432/stashsphere?schema=public";
            };
          };
        });
    };
}
