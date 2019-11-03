{
  buildGoPackage,
  fetchFromGitHub,
}:

let
in buildGoPackage rec {
  name = "rMAPI-${version}";
  version = "0.0.6";

  goPackagePath = "github.com/juruen/rmapi";

  src = fetchFromGitHub {
    owner = "juruen";
    repo = "rmapi";
    rev = "v${version}";
    sha256 = "0nzlgxakckqxip2ngvrcgj32q70zrmx1q7hhsphc0gilhbl9qa05";
  };

  goDeps = ./deps.nix;

  buildFlags = "--tags release";
}
