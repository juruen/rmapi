{stdenv, buildGoPackage}:

let
in buildGoPackage rec {
  name = "rMAPI";

  goPackagePath = "github.com/juruen/rmapi";

  src = ./.;

  goDeps = ./deps.nix;

  buildFlags = "--tags release";

  doCheck = true;

  meta = with stdenv.lib; {
    homepage = "https://github.com/juruen/rmapi";
    description = "rMAPI is a Go app that allows you to access the ReMarkable Cloud API programmatically";
    license = licenses.gpl3;
    platforms = platforms.unix;
  };
}
