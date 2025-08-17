{
  description = "photos";

  inputs = {
    nixpkgs = {
      type = "github";
      owner = "NixOS";
      repo = "nixpkgs";
      rev = "fc02ee70efb805d3b2865908a13ddd4474557ecf";
    };
    flake-utils = {
      type = "github";
      owner = "numtide";
      repo = "flake-utils";
      rev = "b1d9ab70662946ef0850d488da1c9019f3a9752a";
    };
    pre-commit-hooks = {
      type = "github";
      owner = "cachix";
      repo = "git-hooks.nix";
      rev = "c7012d0c18567c889b948781bc74a501e92275d1";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }@inputs:
    let
      utils = flake-utils;
    in
    utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config = {
            allowUnfree = true;
          };
        };

      in
      {
        checks = {
          pre-commit-check = inputs.pre-commit-hooks.lib.${system}.run {
            src = ./.;
            hooks = {
              dprint = {
                enable = true;
                name = "dprint check";
                entry = "dprint check --allow-no-files";
              };
              nixfmt = {
                enable = true;
                name = "nixfmt check";
                entry = "nixfmt -c ";
                types = [ "nix" ];
              };
            };
          };
        };

        formatter = pkgs.nixpkgs-fmt;

        packages.default = pkgs.buildGoModule {
          pname = "photos";
          version = "0.1.0";
          vendorHash = "sha256-Q7V4Kp2t9skOvYEYEj+okWgLFn12fhQWNorn6svs+eA=";
          src = ./.;
          checkPhase = "";
          nativeBuildInputs = with pkgs; [ pkg-config ];
          buildInputs = with pkgs; [
            pkg-config
          ];
        };

        devShells = {
          default = pkgs.mkShell {
            inherit (self.checks.${system}.pre-commit-check) shellHook;
            buildInputs = self.checks.${system}.pre-commit-check.enabledPackages;

            packages = with pkgs; [
              go
              exiftool
              dprint
              nixfmt-rfc-style
              nixfmt-tree
              claude-code
              gci
              golangci-lint
              gofumpt
              google-cloud-sdk
              gotools
              google-cloud-sql-proxy

              postgresql_16 # psql
            ];
          };
        };
      }
    );
}
