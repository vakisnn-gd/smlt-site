#!/bin/sh
set -eu

if [ "$#" -ne 2 ] || { [ "$1" != chat ] && [ "$1" != max ]; }; then
  echo "Usage: $0 chat|max DESTINATION" >&2
  exit 2
fi

app=$1
dest=$2
src=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
case "$dest" in /var/www/*|/opt/chat-smlt/dist/*|/tmp/*) ;; *) echo "Refusing unexpected output path: $dest" >&2; exit 2;; esac

install -d -m 755 "$dest" "$dest/avatars"
install -m 644 "$src/$app/index.html" "$src/$app/app.js" "$src/$app/style.css" "$dest/"
install -m 644 "$src/shared/public/smlt-max-fixed.png" "$dest/"
find "$src/shared/public/avatars" -maxdepth 1 -type f -exec install -m 644 {} "$dest/avatars/" \;
find "$dest" -type d -exec chmod 755 {} +
find "$dest" -type f -exec chmod 644 {} +
