#!/bin/bash

set -euxo pipefail

function check_equal() {
    if [ "$1" != "$2" ]; then
        ink -r "Expected ciux output ($1) to equal $2"
    else
        ink -g "Correct ciux output: $2"
    fi
}

# Run the e2e tests

git_dir=$(mktemp -d)
git init "$git_dir"

mkdir -p "$git_dir/rootfs"
cd $git_dir

# Add a file to the rootfs
file="$git_dir/rootfs/hello.txt"
echo "Hello World" > "$file"
git add "$file"
git commit -m "Add $file"

git_tag="v1.0.0"
git tag -a "$git_tag" -m "Release $git_tag"

ver=$(ciux get imagetag "$git_dir")
check_equal "$ver" "$git_tag"

file="$git_dir/hello2.txt"
echo "Hello World" > "$file"
git add "$file"
git commit -m "Add $file"

ver=$(ciux get imagetag "$git_dir")
check_equal "$ver" "$git_tag"

file="$git_dir/rootfs/hello3.txt"
echo "Hello World" > "$file"
git add "$file"
git commit -m "Add $file"

ver=$(ciux get imagetag "$git_dir")
check_equal "$ver" "$git_tag-2-g$(git rev-parse --short HEAD)"