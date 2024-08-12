#!/bin/bash

set -euxo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. $DIR/include.sh

$DIR/create_git_repo.sh

img_url=$(ciux get image "$git_dir")
expected_img_url="Image: test_url/test_org/$project:$git_tag, in registry: false"
check_equal "$expected_img_url" "$img_url"

file="$git_dir/hello2.txt"
ink "Commit $file"
echo "Hello World" > "$file"
git add "$file"
git commit -m "Add $file"

img_url=$(ciux get image "$git_dir")
expected_img_tag="$git_tag-1-g$(git rev-parse --short HEAD)"
expected_img_url="Image: test_url/test_org/$project:$expected_img_tag, in registry: false"
check_equal "$expected_img_url" "$img_url"

file="$git_dir/rootfs/hello3.txt"
ink "Commit $file"
echo "Hello World" > "$file"
git add "$file"
git commit -m "Add $file"

img_url=$(ciux get image "$git_dir")
expected_img_tag="$git_tag-2-g$(git rev-parse --short HEAD)"
expected_img_url="Image: test_url/test_org/$project:$expected_img_tag, in registry: false"
check_equal "$expected_img_url" "$img_url"

ink "Check image url"
if img_url=$(ciux get image --check "$git_dir")
then
    check_equal "$expected_img_url" "$img_url"
fi

