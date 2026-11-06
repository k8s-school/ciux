#!/bin/bash

# Check that 'ciux get image --check' rebuilds an image whose base image,
# a dependency labelled "build" in .ciux, changed without any source change

set -euxo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. $DIR/include.sh

registry_port=5000
registry="localhost:$registry_port"
registry_container="ciux-e2e-registry"

docker rm -f "$registry_container" >/dev/null 2>&1 || true
docker run -d --name "$registry_container" -p "$registry_port:5000" registry:2
trap 'docker rm -f "$registry_container" >/dev/null 2>&1 || true' EXIT

# Build and push the two versions of the base image
for version in v1 v2; do
    base_dir="$tmp_dir/base-$version"
    mkdir -p "$base_dir"
    cat > "$base_dir/Dockerfile" <<EOF
FROM busybox:1.36
RUN echo "base $version" > /base-version
EOF
    docker build -t "$registry/base:$version" "$base_dir"
    docker push "$registry/base:$version"
done

project="get_image_base"
git_dir="$tmp_dir/$project"
mkdir -p "$git_dir/rootfs"
cd "$git_dir"
git init
git config user.email "you@example.com"
git config user.name "Your Name"

function write_ciux_config() {
    cat > "$git_dir/.ciux" <<EOF
apiVersion: v1alpha1
registry: $registry
sourcePathes:
  - Dockerfile
  - rootfs
dependencies:
  - image: $registry/base:$1
    labels:
      build: ""
EOF
}

write_ciux_config v1
cat > "$git_dir/Dockerfile" <<EOF
ARG base_image
FROM \${base_image}
COPY rootfs /rootfs
EOF
echo "Hello World" > "$git_dir/rootfs/hello.txt"
git add .
git commit -m "Add project built on base v1"
git tag -a "$git_tag_v1" -m "Release $git_tag_v1"

# Build and push the project image, as a CI would do, if ciux asks for it
function build_if_needed() {
    $(ciux get image --check "$git_dir" --env)
    if [ "$CIUX_BUILD" = true ]; then
        docker build -t "$CIUX_IMAGE_URL" --build-arg base_image="$registry/base:$1" "$git_dir"
        docker push "$CIUX_IMAGE_URL"
    fi
}

ink "Image built on base v1"
build_if_needed v1
img_url=$(ciux get image --check "$git_dir")
check_equal "$img_url" "Image: $registry/$project:$git_tag_v1, in registry: true"

ink "Bump the base image to v2 in .ciux, without any source change"
write_ciux_config v2
git commit -am "Use base v2"
bump_tag="${git_tag_v1}-1-g$(git rev-parse --short HEAD)"
img_url=$(ciux get image --check "$git_dir")
check_equal "$img_url" "Image: $registry/$project:$bump_tag, in registry: false"

ink "Image rebuilt on base v2"
build_if_needed v2
img_url=$(ciux get image --check "$git_dir")
check_equal "$img_url" "Image: $registry/$project:$bump_tag, in registry: true"
