# Ciux Documentation

## Table of Contents

- [Ciux Documentation](#ciux-documentation)
  - [Table of Contents](#table-of-contents)
  - [1. Introduction](#1-introduction)
    - [Overview](#overview)
    - [Key Features](#key-features)
  - [2. Getting Started](#2-getting-started)
    - [Installation](#installation)
    - [Configuration](#configuration)
  - [3. Usage](#3-usage)
    - [Prerequisites](#prerequisites)
    - [Building a simple project with ciux:](#building-a-simple-project-with-ciux)
    - [Integration Tests](#integration-tests)
    - [Building a multi-repository project with ciux:](#building-a-multi-repository-project-with-ciux)

## 1. Introduction

### Overview

Ciux is a versatile tool designed to automate build processes and streamline integration testing. This documentation provides guidance on installing, configuring, and utilizing Ciux in your development workflows.

### Key Features

- **Build Automation:** Ciux automates the build process, ensuring consistency and reliability in source code dependencies generating software artifacts.

- **Integration Testing:** Facilitates seamless integration testing with support for various dependencies and configurations.

## 2. Getting Started

### Installation

To install Ciux, use the following command:

```bash
$ go install github.com/k8s-school/ciux@<version>
```

### Configuration

Ciux requires a configuration file (`.ciux`) at the top level of your source code repository. Example configuration:

```yaml
apiVersion: v1alpha1
registry: gitlab-registry.in2p3.fr/astrolabsoftware/fink
dependencies:
  - url: https://github.com/astrolabsoftware/fink-alert-simulator
    clone: true
    pull: true
    labels:
      itest: "true"
      ci: "true"
  - url: https://github.com/astrolabsoftware/finkctl
    clone: true
    labels:
      itest: "true"
      ci: "true"
  - image: gitlab-registry.in2p3.fr/astrolabsoftware/fink/spark-py:k8s-3.4.1
    labels:
      build: "true"
  - package: github.com/k8s-school/ktbx@v1.1.1-rc7
    labels:
      itest: "optional"
      ci: "true"
```

## 3. Usage

### Prerequisites

1. The `CIUXCONFIG` Variable

This variable store the path to the current project configuration file produced by `ciux`, if undefined, `ciux` will store this file in `<PROJECT_DIR>/.ciux.d`.

Before using Ciux, the `CIUXCONFIG` variable can be defined.
```bash
$ export CIUXCONFIG="$HOME/.ciux/ciux.sh"
```

- The `CIUXCONFIG` variable is utilized by `ciux` to dynamically retrieve source code and version information during the build and integration testing processes.

### Building a simple project with ciux:

1. Prepare the build process by generating the `CIUXCONFIG` file, using `ciux ignite`:

```bash
$ cd <project-source-directory>
$ ciux ignite --selector build
```

2. Example `CIUXCONFIG` content:

```bash
export FINK_BROKER_DIR=/home/fjammes/src/astrolabsoftware/fink-broker
export FINK_BROKER_VERSION=v3.1.1-rc1-7-ga4bf010
export ASTROLABSOFTWARE_FINK_SPARK_PY_IMAGE=gitlab-registry.in2p3.fr/astrolabsoftware/fink/spark-py:k8s-3.4.1
```

There is a `CIUXCONFIG` generate by `ciux ignite` for each different label selector. This allow for example to have a different `CIUXCONFIG` for the build or the integration test

3. Explanation:

   - The `ciux ignite` command orchestrates the following crucial steps:

     1. **Dependency Selection:** This phase involves dynamic selection of dependencies based on specified labels such as `build`. Ciux identifies and includes only those dependencies that are relevant to the current workflow.

     2. **Image Existence Check:** Ciux ensures the availability of images in the remote registry associated with the specified dependencies. This verification step guarantees that the required images are present and accessible before initiating subsequent processes.

     3. **Git Cloning Process:**
         - For each selected dependency, Ciux performs a Git clone operation, fetching the source code from the respective repositories. The cloning process use the very same branch of the main project. If the specified branch does not exist if a dependency, Ciux defaults to `main` or `master`.

     4. **Go Package Installation:** Additionally, during the ignite process, Ciux installs Go packages that are essential for the project and its dependencies. This ensures that the required Go packages are available and compatible with the project's build and integration testing.

   - The generated `CIUXCONFIG` file encapsulates the dynamically determined information about source code locations, versions, and package installations. This variable is crucial for Ciux to maintain a consistent and reproducible development environment during subsequent build and integration processes.

   - Utilizing the `CIUXCONFIG` variables, Ciux facilitates a streamlined workflow, automating the setup of dependencies and environment configurations essential for successful integration testing and continuous integration.

4. Build the project

   - To build the project using Ciux, a two-step process is required. It involves sourcing the `CIUXCONFIG` file, which sets essential environment variables, and subsequently running the project's build script that relies on these environment variables.

     1. **Source `CIUXCONFIG` to Set Environment Variables:**

     Prior to initiating the build process, source the `CIUXCONFIG` file to set the necessary environment variables. This file contains crucial information about the project's dependencies, source code locations, and version details.

     ```bash
      # Be careful to use the same label selector than the one used for `ciux ignite`
      if ciuxconfig=$(ciux get configpath --selector "build" "$git_dir" 2>&1); then
        source "$ciuxconfig"
      else
        echo "Error while loading ciux config : $ciuxconfig" >&2
        exit 1
      fi
     ```

   - Sourcing `CIUXCONFIG` ensures that the environment variables required for the build process are properly configured and accessible.

5. **Run the Build Script of the Project:**
   - With the environment variables set by sourcing `CIUXCONFIG`, execute the build script of the project. This script is responsible for compiling, assembling, and generating the project artifacts.

     ```bash
     $  <project-source-directory>/path/to/build/script.sh
     ```

   - The build script should be designed to utilize the environment variables set by Ciux, such as source code locations, version details, and any other configurations specified in the `CIUXCONFIG` file.


### Integration Tests

1. Prepare integration tests by generating the `CIUXCONFIG` file, using `ciux ignite` with dedicated labels:

```bash
$ cd <source-directory>
$ ciux ignite --selector ci
```

2. Example `CIUXCONFIG` content:

```bash
export FINK_ALERT_SIMULATOR_DIR=/home/fjammes/src/astrolabsoftware/fink-alert-simulator
export FINK_ALERT_SIMULATOR_VERSION=v3.1.1-rc1
export FINKCTL_DIR=/home/fjammes/src/astrolabsoftware/finkctl
export FINKCTL_VERSION=v3.1.1-rc2-2-g68a1d41
export FINK_BROKER_DIR=/home/fjammes/src/astrolabsoftware/fink-broker
export FINK_BROKER_VERSION=v3.1.1-rc1-7-ga4bf010
```

3. Run integration tests

```bash
$ source $CIUXCONFIG
 <project-source-directory>/path/to/integration-test/script.sh
```

### Building a multi-repository project with ciux:

When working on a project that is split across multiple Git repositories, it's important to understand how a Git-based continuous integration (CI) system will determine which branches to use when building the project.

If you have created a branch with the same name (e.g. `my-feature-branch`) in all of the repositories, then the `ciux` will use that branch for the build. This is because the `ciux` will look for a branch with the same name in each repository and use that branch for the build.

However, if you have not created a branch with the same name in all of the repositories, then the `ciux` will use the `main` or `master` branch of the other repositories. This is because `ciux` needs to have a common baseline to build against, and if there is no branch with the same name in all repositories, then it will default to using the `main` or `master` branch.

Here's an example of how this might work in practice:

Suppose you have a project that consists of three repositories: `repo1`, `repo2`, and `repo3`. You want to create a new feature branch called `my-feature-branch` in all three repositories.

To do this, you would first create the branch in `repo1`:
```bash
git checkout -b my-feature-branch
```
Then, you would push the branch to the remote repository:
```bash
git push -u origin my-feature-branch
```
You would then repeat this process for `repo2` and `repo3`.

Now, when you trigger a build in the CI system, it will look for a branch called `my-feature-branch` in all three repositories and use that branch for the build.

However, if you had only created the `my-feature-branch` branch in `repo1` and `repo2`, but not in `repo3`, then the CI system would use the `main` or `master` branch of `repo3` for the build. This is because the CI system needs to have a common baseline to build against, and if there is no branch with the same name in all repositories, then it will default to using the `main` or `master` branch.