{
  description = "Terraform Module Version Management Tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
        # Version should match .version file (sync before releases)
        version = "0.1.0";
      in {
        packages.default = pkgs.buildGoModule {
          pname = "terraform-module-versions";
          inherit version;

          src = self;

          vendorHash = "sha256-SycyR6HaS9zpfrgxSZel0s8EHRkWcdcAixtdefkp3Ug=";

          subPackages = ["cmd/tf-update-module-versions"];

          ldflags = [
            "-X main.Version=${version}"
          ];

          doCheck = true;

          checkFlags = [
            "-v" # verbose output
          ];

          meta = {
            description = "Discover, analyze, and update Terraform module versions";
            homepage = "https://github.com/vdesjardins/terraform-module-versions";
            license = pkgs.lib.licenses.mit;
            mainProgram = "tf-update-module-versions";
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go compiler
            go_1_25

            # Development tools
            golangci-lint
            gnumake
            git
          ];

          shellHook = ''
            echo "terraform-module-versions development environment loaded"
            echo "Available commands: make build, make test, make lint, make fmt, make clean"
            echo "Run 'make help' for more information"
          '';
        };
      }
    );
}
