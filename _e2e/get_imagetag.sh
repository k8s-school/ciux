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

tmp_dir=$(mktemp -d)
project="e2e"
git_dir="$tmp_dir/$project"
mkdir "$git_dir"
cd "$git_dir"
git init "$git_dir"

mkdir -p "$git_dir/rootfs"

# Add ciux config file
file="$git_dir/.ciux"

cat > "$file" <<EOF
apiVersion: v1alpha1
registry: test_url/test_org
EOF
git add "$file"
git commit -m "Add $file"

# Add a file to the rootfs
file="$git_dir/rootfs/hello.txt"
echo "Hello World" > "$file"
git add "$file"
git commit -m "Add $file"

git_tag="v1.0.0"
git tag -a "$git_tag" -m "Release $git_tag"

ver=$(ciux get imagetag "$git_dir")
check_equal "$git_tag" "$ver"

file="$git_dir/hello2.txt"
echo "Hello World" > "$file"
git add "$file"
git commit -m "Add $file"

ver=$(ciux get imagetag "$git_dir")
check_equal "$git_tag" "$ver"

file="$git_dir/rootfs/hello3.txt"
echo "Hello World" > "$file"
git add "$file"
git commit -m "Add $file"

img_tag=$(ciux get imagetag "$git_dir")
expected_img_tag="$git_tag-2-g$(git rev-parse --short HEAD)"
check_equal "$expected_img_tag" "$img_tag"

img_url=$(ciux get imagetag --full "$git_dir")
expected_img_url="test_url/test_org/$project:$expected_img_tag"
check_equal "$expected_img_url" "$img_url"