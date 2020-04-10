set -e
path=$(dirname $0)
go test -v github.com/juruen/rmapi/annotations 
xdg-open /tmp/a3.pdf
xdg-open /tmp/a4.pdf
xdg-open /tmp/a5.pdf
xdg-open /tmp/rm.pdf
xdg-open /tmp/letter.pdf
