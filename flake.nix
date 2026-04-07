{
  description = "Terminal screensaver and ambient visualiser";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = self.shortRev or self.dirtyShortRev or "dev";
      in {
        packages.default = pkgs.buildGoModule {
          pname = "drift";
          inherit version;
          src = ./.;

          # Run `nix build` once; replace with the hash from the error:
          #   error: hash mismatch in fixed-output derivation ... got: sha256-...
          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";

          CGO_ENABLED = 0;

          ldflags = [
            "-s"
            "-w"
            "-X main.version=${version}"
            "-X main.commit=${self.rev or "none"}"
            "-X main.date=unknown"
          ];

          meta = with pkgs.lib; {
            description = "Terminal screensaver and ambient visualiser";
            homepage = "https://github.com/phlx0/drift";
            license = licenses.mit;
            maintainers = [ ];
            mainProgram = "drift";
          };
        };

        apps.default = flake-utils.lib.mkApp {
          drv = self.packages.${system}.default;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [ go gopls golangci-lint ];
        };
      });
}
