{ pkgs ? import <nixpkgs> {} }:
with pkgs;

let
in pkgs.callPackage ./derivation.nix {
  buildGoModule = super: buildGoModule (super // {

    name = super.name + "-env";

    buildInputs = (super.buildInputs or []) ++ [
      # Project dependencies

      # utilities
      git
    ];

  });
}
