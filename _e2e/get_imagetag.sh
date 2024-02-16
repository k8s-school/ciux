#!/bin/bash

set -euxo pipefail

function check_equal() {
    if [ "$1" != "$2" ]; then
        ink -r "Expected ciux output ($1) to equal $2"
	exit 1
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
git config --global user.email "you@example.com"
git config --global user.name "Your Name"

mkdir -p "$git_dir/rootfs"

# Add ciux config file
file="$git_dir/.ciux"

cat > "$file" <<EOF
apiVersion: v1alpha1
registry: test_url/test_org
sourcePathes:
  - rootfs
EOF
git add "$file"
git commit -m "Add $file"
ciux ignite "$git_dir" --selector "itest=true"


file="$git_dir/rootfs/hello.txt"
ink "Commit $file"
echo "Hello World" > "$file"
git add "$file"
git commit -m "Add $file"

git_tag="v1.0.0"
ink "Tag $git_tag"
git tag -a "$git_tag" -m "Release $git_tag"

img_url=$(ciux get image "$git_dir")
expected_img_url="test_url/test_org/$project:$git_tag"
check_equal "$expected_img_url" "$img_url"

file="$git_dir/hello2.txt"
ink "Commit $file"
echo "Hello World" > "$file"
git add "$file"
git commit -m "Add $file"

img_url=$(ciux get image "$git_dir")
check_equal "$expected_img_url" "$img_url"

file="$git_dir/rootfs/hello3.txt"
ink "Commit $file"
echo "Hello World" > "$file"
git add "$file"
git commit -m "Add $file"

img_url=$(ciux get image "$git_dir")
expected_img_tag="$git_tag-2-g$(git rev-parse --short HEAD)"
expected_img_url="test_url/test_org/$project:$expected_img_tag"
check_equal "$expected_img_url" "$img_url"

ink "Check image url"
if img_url=$(ciux get image --check "$git_dir")
then
    ink -r "ciux get image --full failed"
else
    check_equal "$expected_img_url" "$img_url"
fi

