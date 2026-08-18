{
  description = "aps - AWS profile switcher";

  inputs = {
    nixpkgs.url = "nixpkgs/nixpkgs-26.05-darwin";
  };

  outputs = {
    self,
    nixpkgs,
  }: let
    systems = ["aarch64-darwin" "x86_64-darwin" "aarch64-linux" "x86_64-linux"];
    forAllSystems = nixpkgs.lib.genAttrs systems;
    pkgsFor = system: import nixpkgs {inherit system;};
  in {
    packages = forAllSystems (system: let
      pkgs = pkgsFor system;
    in {
      default = pkgs.buildGoModule {
        pname = "aps";
        version = "0.1.0";
        src = ./.;
        vendorHash = "sha256-TiOg0XL2I0KavA0s1eBVW2mmR6MZoKnnGLD6iD9iY1U=";

        meta = {
          description = "AWS profile switcher";
          homepage = "https://github.com/obvionaoe/aps";
          mainProgram = "aps";
          platforms = pkgs.lib.platforms.unix;
        };
      };
    });

    overlays.default = final: prev: {
      aps = self.packages.${final.system}.default;
    };

    devShells = forAllSystems (system: let
      pkgs = pkgsFor system;
    in {
      default = pkgs.mkShell {
        packages = [pkgs.go pkgs.gofumpt];
      };
    });
  };
}
