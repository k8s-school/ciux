#!/bin/bash

set -euxo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. $DIR/include.sh

# Create a test project directory
project_dir="$tmp_dir/revision_pin_test"
mkdir -p "$project_dir"

# Create a dependency repository
dep_dir="$tmp_dir/test_dependency"
$DIR/create_git_repo.sh "$dep_dir"

cd "$dep_dir"

# Create additional commits and tags in the dependency
echo "Second file" > second.txt
git add second.txt
git commit -m "Add second file"
git tag v2.0.0

echo "Third file" > third.txt
git add third.txt
git commit -m "Add third file"
third_commit=$(git rev-parse HEAD)

cd "$project_dir"

# Initialize project repository
git init
echo "Project file" > project.txt
git add project.txt
git commit -m "Initial project commit"

# Create .ciux configuration with pinned revision
cat > .ciux << EOF
apiVersion: v1alpha1
registry: test-registry.io
sourcePathes:
  - project.txt
dependencies:
  - url: file://$dep_dir
    clone: true
    revision: v1.0.0
    labels:
      test: "true"
  - url: file://$dep_dir
    clone: true
    revision: $third_commit
    labels:
      build: "true"
EOF

ink "Test ignite with pinned revision (tag)"
ciux ignite --selector="test=true" "$project_dir"

# Check that the dependency was cloned to the correct revision (v1.0.0)
test_dep_dir="$tmp_dir/test_dependency"
if [ -d "$test_dep_dir" ]; then
    cd "$test_dep_dir"
    current_commit=$(git rev-parse HEAD)
    v1_commit=$(git rev-parse v1.0.0)
    check_equal "$current_commit" "$v1_commit"
    ink -g "Dependency correctly checked out to v1.0.0"

    # Verify that second.txt and third.txt don't exist (since we're on v1.0.0)
    if [ -f "second.txt" ] || [ -f "third.txt" ]; then
        ink -r "ERROR: Files from later commits found in v1.0.0 checkout"
        exit 1
    fi
    ink -g "Verified that later commits are not present in v1.0.0 checkout"
fi

cd "$project_dir"

ink "Test ignite with pinned revision (commit hash)"
# Clean up previous clone
rm -rf "$test_dep_dir"

ciux ignite --selector="build=true" "$project_dir"

# Check that the dependency was cloned to the correct revision (third commit)
if [ -d "$test_dep_dir" ]; then
    cd "$test_dep_dir"
    current_commit=$(git rev-parse HEAD)
    check_equal "$current_commit" "$third_commit"
    ink -g "Dependency correctly checked out to commit $third_commit"

    # Verify that third.txt exists (since we're on the third commit)
    if [ ! -f "third.txt" ]; then
        ink -r "ERROR: third.txt not found in commit $third_commit checkout"
        exit 1
    fi
    ink -g "Verified that third.txt exists in commit $third_commit checkout"
fi

ink -g "All revision pinning tests passed!"