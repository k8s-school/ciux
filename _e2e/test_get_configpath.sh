#!/bin/bash

set -euxo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. $DIR/include.sh

git_dir="$tmp_dir/get_configpath"
$DIR/create_git_repo.sh "$git_dir"

ciux ignite -l "e2e" "$git_dir"

if ciuxconfig=$(ciux get configpath -l "e2e" "$git_dir" 2>&1); then
    source "$ciuxconfig"
else
    echo "Error while loading ciux config : $ciuxconfig" >&2
    exit 1
fi

expected_ciuxconfig="$tmp_dir/get_configpath/.ciux.d/ciux_e2e.sh"
check_equal "$expected_ciuxconfig" "$ciuxconfig"

expected_ciuximageurl="test_url/test_org/get_configpath:$git_tag_v1"
check_equal "$expected_ciuximageurl" "$CIUX_IMAGE_URL"

