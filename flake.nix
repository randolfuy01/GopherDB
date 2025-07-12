{
  description = "Go project development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        # Define your Go application
        goApp = pkgs.buildGoModule {
          pname = "my-go-app";
          version = "0.1.0";
          
          src = ./.;
          
          vendorHash = null;
          
          meta = with pkgs.lib; {
            description = "Gopher-DB env";
            license = licenses.mit;
            maintainers = [ ];
          };
        };
        
      in
      {
        # Development environment
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls          # Go language server
            gotools        # Go tools (goimports, etc.)
            go-migrate     # Database migrations (optional)
            air            # Live reload for Go apps (optional)
          ];
          
          shellHook = ''
            echo "Go development environment loaded"
            echo "Go version: $(go version)"
            echo ""
            echo "Available commands:"
            echo "  go run .           - Run the application"
            echo "  go build .         - Build the application"
            echo "  go test ./...      - Run tests"
            echo "  air                - Run with live reload (if air is used)"
            echo "  nix run            - Run the built application"
            echo "  nix build          - Build the application with Nix"
          '';
        };
        
        # Default package
        packages.default = goApp;
        
        # Run the application
        apps.default = {
          type = "app";
          program = "${goApp}/bin/my-go-app";
        };
        
        # Additional apps for development
        apps.dev = {
          type = "app";
          program = "${pkgs.writeShellScript "dev-server" ''
            echo "Starting Go development server..."
            ${pkgs.go}/bin/go run .
          ''}";
        };
        
        apps.watch = {
          type = "app";
          program = "${pkgs.writeShellScript "watch-server" ''
            echo "Starting Go server with live reload..."
            ${pkgs.air}/bin/air
          ''}";
        };
      });
}
