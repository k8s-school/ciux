# Feature Documentation: Improved Image Pulling Logic

## Overview:
In the previous version, the process involved searching for the latest commit (referred to as C1) where there was a code modification in the image. Subsequently, pulling the associated image from the registry was conducted. However, this triggered a rebuild if the image did not exist. The current enhancement involves scanning all the commits between the current commit and C1 to determine if an image has been built with the current code. This optimization, albeit minor, aims to mitigate failures in certain builds within the CI pipeline.

## Changes:
Image Pulling Logic Enhancement:
Previously: Only the latest commit with a code modification triggered image pulling.
Now: Scanning all commits between the current commit and C1 to identify images built with the current code.
Reduced Build Failures:
By scanning for images built with the current code, the probability of triggering unnecessary rebuilds due to missing images in the registry is significantly reduced.
Implementation Details:
The process involves iterating through commits between the current commit and C1.
For each commit:
Check if an image has been built with the current code.
If found, pull the associated image from the registry.
If not found, proceed to the next commit.
This logic is integrated into the image pulling module of the system.
Impact:
Improved CI Stability: By avoiding unnecessary rebuilds triggered by missing images, the stability of Continuous Integration (CI) pipelines is enhanced.
Minor Optimization: While the impact may seem minor, it addresses specific failure scenarios in the CI environment.