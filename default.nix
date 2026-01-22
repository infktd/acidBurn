# This file provides backward compatibility for non-flake users
# Usage: nix-build
(builtins.getFlake (toString ./.)).packages.${builtins.currentSystem}.default
