{
  description = "A terminal dashboard for managing devenv.sh projects";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = {
          default = pkgs.buildGoModule {
            pname = "devdash";
            version = "0.1.0";

            src = ./.;

            vendorHash = "sha256-CwWLZBudC+cUEDWZAqhOhjwLYybjATrC9VG7GNVLNnQ=";

            ldflags = [
              "-s"
              "-w"
            ];

            meta = with pkgs.lib; {
              description = "A terminal dashboard for managing devenv.sh projects";
              homepage = "https://github.com/infktd/devdash";
              license = licenses.mit; # Change if different
              maintainers = [ ];
              mainProgram = "devdash";
            };
          };
        };

        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/devdash";
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_25
            gopls
          ];
        };
      }
    );
}
