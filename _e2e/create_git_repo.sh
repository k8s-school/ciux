#!/bin/bash

set -euxo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Create a git repo
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

ink "Tag $git_tag_v1"
git tag -a "$git_tag_v1" -m "Release $git_tag_v1"
