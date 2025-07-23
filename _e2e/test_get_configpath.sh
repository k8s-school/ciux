#!/bin/bash

set -euxo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. $DIR/include.sh

git_dir="$tmp_dir/get_configpath"
$DIR/create_git_repo.sh "$git_dir"

SELECTOR="e2e"
PROJECT_DIR="$git_dir"
ciux ignite -l "$SELECTOR" "$git_dir"

. $git_dir/.ciux.d/ciuxconfig.sh

expected_ciuxconfig="$tmp_dir/get_configpath/.ciux.d/ciux_${SELECTOR}.sh"
check_equal "$expected_ciuxconfig" "$ciuxconfig"

expected_ciuximageurl="test_url/test_org/get_configpath:$git_tag_v1"
check_equal "$expected_ciuximageurl" "$CIUX_IMAGE_URL"

