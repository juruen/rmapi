## rmapi 0.0.6 (September 08, 2019)

* Migrate to go11 modules

* Add api.v2

## rmapi 0.0.5 (August 11, 2019)

* Fix issue to make document uploads work with reMarkable 1.7.x firmware upgrade

* Increased http timeout to 5 minutes to enable upload of larger files

* Add user-agent header to be a good reMarkable citezen

* Ls may take a directory as an argument

* Ignore hidden files/directories by default

* Initial support for annotations

* Fix panic when autocompleting the "ls" command

* Add find command

## rmapi 0.0.4 (October 1, 2018)

* Windows fixes

* Add autocompleter for local files that is used by "put"

* Fix mv command

* Put may take a second argument as destination

* Autocomplete for "put" command only shows ".pdf" files

* Add support to upload epub files

* rm supports multiple files

* Return exit code for non-interactivly commands

* Vendorize fuse dependencies

* Use new auth endpoints

## rmapi 0.0.3 (February 25, 2018)

* Update doc

* Fix file upload

   *Javier Uruen Val*

## rmapi 0.0.2 (February 25, 2018)

*  Fix directory creation (fixes #6)

*  Add stat command to show entry's metadata

   *Javier Uruen Val*

## rmapi 0.0.1 (February 24, 2018)

*   Initial release with support for most of the API and autocompletion.

    *Javier Uruen Val*
