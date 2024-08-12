#!/bin/bash

set -euxo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. $DIR/include.sh

$DIR/create_git_repo.sh

cd "$git_dir"

expected_tag="$git_tag_v1"

ink "Check release"
release=$(ciux get revision --isrelease "$git_dir")
check_equal "$expected_tag" "$release"

ink "Check is not release"
file="$git_dir/hello2.txt"
ink "Commit $file"
echo "Hello World" > "$file"
git add "$file"
git commit -m "Add $file"
expected_tag=""
release=$(ciux get revision --isrelease "$git_dir")
check_equal "$expected_tag" "$release"


