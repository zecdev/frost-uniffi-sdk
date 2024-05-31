# How to make a release of this repo

Currently the UniFFI repo only generates swift bindings.


To generate a release the FFI code for swift must be generated and
built into an Xcframework that will be published as a github release


**No binaries should be commited to the repo**


## Using CI
**pre-requisites**: you need to be a repo maintainer


1. as a *maintainer* pull the the latest changes from main.
2. branch to a new release. Example: to create release 0.0.1
create a new branch called `release-0.0.1` using `git checkout -b release-0.0.1`
3. the `release.yml` workflow should be executed. this will build
the project and create a github release with the version number provided
in the branch name, containing the artifacts and generated release notes 
by making a diff of the commits from the latest tag to this one.


## Manual release

**pre-requisites**: you need a macOS computer with Xcode installed and
be a repo maintainer.


1. as a *maintainer* create a branch with any name.
2. run `sh Scripts/build_swift.sh`. the script will create the xcframework,
sync the resulting files with the FrostSwiftFFI folder, package it in a zip and 
calculate the checksum
3. update `Package.swift` with the checksum and the version number you wish to 
release in the `url:` part of the`binarytarget` of the package.
4. commit the changes. **No binaries should be commited to the repo**
5. tag the release (e.g: git tag 0.0.1) and push the tag
6. manually create a new Github release from the tag you pushed.
7. in the release include the binary that the Package.swift will be
referencing.
8. **VERIFY THAT THE RESULTING URL POINTS TO THE RIGHT ARTIFACT**
9. **VERIFY THAT THE ARTIFACT DOWNLOADED BY THE `Package.swift` FILE HAS THE THE RIGHT CHECKSUM**

