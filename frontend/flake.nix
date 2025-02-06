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
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          # dev shell frontend
          default = pkgs.mkShell {
            name = "default";
            buildInputs = with pkgs; [
              nodePackages.npm
              nodejs_22
              pnpm
            ];
          };
        });
    };
}
