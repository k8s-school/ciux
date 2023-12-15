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
    - [Building a project with ciux:](#building-a-project-with-ciux)
    - [Integration Tests](#integration-tests)

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

1. Defining the `CIUXCONFIG` Variable

Before using Ciux, the `CIUXCONFIG` variable must be defined.
```bash
$ export CIUXCONFIG="$HOME/.ciux/ciux.sh"
```

- The `CIUXCONFIG` variable is utilized by Ciux to dynamically retrieve source code and version information during the build and integration testing processes.

### Building a project with ciux:

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

3. Explanation:

   - The `ciux ignite` command orchestrates the following crucial steps:

     1. **Dependency Selection:** This phase involves dynamic selection of dependencies based on specified labels such as `build`. Ciux identifies and includes only those dependencies that are relevant to the current workflow.

     2. **Image Existence Check:** Ciux ensures the availability of images in the remote registry associated with the specified dependencies. This verification step guarantees that the required images are present and accessible before initiating subsequent processes.

     3. **Git Cloning Process:**
         - For each selected dependency, Ciux performs a Git clone operation, fetching the source code from the respective repositories. The cloning process use the very same branch of the main project. If the specified branch does not exist if a dependency, Ciux defaults to `main` or `master`.

     4. **Go Package Installation:** Additionally, during the ignite process, Ciux installs Go packages that are essential for the project and its dependencies. This ensures that the required Go packages are available and compatible with the project's build and integration testing.

   - The generated `CIUXCONFIG` variable encapsulates the dynamically determined information about source code locations, versions, and package installations. This variable is crucial for Ciux to maintain a consistent and reproducible development environment during subsequent build and integration processes.

   - Utilizing the `CIUXCONFIG` variable, Ciux facilitates a streamlined workflow, automating the setup of dependencies and environment configurations essential for successful integration testing and continuous integration.

4. Build the project

To build the project using Ciux, a two-step process is required. It involves sourcing the `CIUXCONFIG` file, which sets essential environment variables, and subsequently running the project's build script that relies on these environment variables.

1. **Source `CIUXCONFIG` to Set Environment Variables:**
   - Prior to initiating the build process, source the `CIUXCONFIG` file to set the necessary environment variables. This file contains crucial information about the project's dependencies, source code locations, and version details.

     ```bash
     $ source $CIUXCONFIG
     ```

   - Sourcing `CIUXCONFIG` ensures that the environment variables required for the build process are properly configured and accessible.

2. **Run the Build Script of the Project:**
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