#!/bin/bash

# Check that 'ciux get image --check' rebuilds an image whose base image,
# the dependency with "baseOf: <build>" in .ciux, changed without any source
# change, and that it gives this base image to the build in CIUX_BASE_IMAGE.
# As in fink-broker, the "science" base is built on the "noscience" one, and
# only the "science" base is bumped: the old image still starts with the
# layers of the current "noscience" base, so only the base of the build must
# be checked. "science" is the default build: its image has no suffix.

set -euxo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. $DIR/include.sh

registry_port=5000
registry="localhost:$registry_port"
registry_container="ciux-e2e-registry"

docker rm -f "$registry_container" >/dev/null 2>&1 || true
docker run -d --name "$registry_container" -p "$registry_port:5000" registry:2
trap 'docker rm -f "$registry_container" >/dev/null 2>&1 || true' EXIT

# Build and push an image from a Dockerfile given on stdin
function build_base() {
    local image="$1"
    local base_dir="$tmp_dir/$(echo "$image" | tr ':/' '__')"
    mkdir -p "$base_dir"
    cat > "$base_dir/Dockerfile"
    docker build -t "$registry/$image" "$base_dir"
    docker push "$registry/$image"
}

build_base noscience:v1 <<EOF
FROM busybox:1.36
RUN echo "noscience v1" > /noscience-version
EOF
for version in v1 v2; do
    build_base science:$version <<EOF
FROM $registry/noscience:v1
RUN echo "science $version" > /science-version
EOF
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
defaultBuild: science
dependencies:
  - image: $registry/noscience:v1
    baseOf: noscience
  - image: $registry/science:$1
    baseOf: science
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
git commit -m "Add project built on science v1"
git tag -a "$git_tag_v1" -m "Release $git_tag_v1"

# Build and push the project image, as a CI would do, if ciux asks for it,
# on the base image given by ciux
function build_if_needed() {
    $(ciux get image --check "$git_dir" --build science --env)
    check_equal "$CIUX_BASE_IMAGE" "$registry/science:$1"
    if [ "$CIUX_BUILD" = true ]; then
        docker build -t "$CIUX_IMAGE_URL" --build-arg base_image="$CIUX_BASE_IMAGE" "$git_dir"
        docker push "$CIUX_IMAGE_URL"
    fi
}

ink "Image built on science v1"
build_if_needed v1
img_url=$(ciux get image --check "$git_dir" --build science)
check_equal "$img_url" "Image: $registry/$project:$git_tag_v1, in registry: true"
img_url=$(ciux get image --check "$git_dir")
check_equal "$img_url" "Image: $registry/$project:$git_tag_v1, in registry: true"

ink "The noscience build has a suffix and is not built yet"
img_url=$(ciux get image --check "$git_dir" --build noscience)
check_equal "$img_url" "Image: $registry/$project-noscience:$git_tag_v1, in registry: false"

ink "Bump the science base image to v2 in .ciux, without any source change"
write_ciux_config v2
git commit -am "Use science v2"
bump_tag="${git_tag_v1}-1-g$(git rev-parse --short HEAD)"
img_url=$(ciux get image --check "$git_dir" --build science)
check_equal "$img_url" "Image: $registry/$project:$bump_tag, in registry: false"

ink "ciux ignite selects the same image to deploy"
ciux ignite --selector itest "$git_dir" --build science
. "$git_dir/.ciux.d/ciux_itest.sh"
check_equal "$CIUX_IMAGE_URL" "$registry/$project:$bump_tag"
check_equal "$CIUX_BASE_IMAGE" "$registry/science:v2"

ink "Image rebuilt on science v2"
build_if_needed v2
img_url=$(ciux get image --check "$git_dir" --build science)
check_equal "$img_url" "Image: $registry/$project:$bump_tag, in registry: true"
