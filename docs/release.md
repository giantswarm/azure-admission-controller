# How to publish a release

TL;DR: Releases are published automatically whenever a new tag of the format X.Y.Z is pushed to the GitHub repository.

## Create a tag and release draft

All you have to do is create and push a new tag.

In the command below, replace `<MAJOR.MINOR.PATCH>` with the actual version number you want to publish.

```
export VERSION=<MAJOR.MINOR.PATCH>
git checkout master
git pull
git tag -a ${VERSION} -m "Release version ${VERSION}"
git push origin ${VERSION}
```

Follow CircleCI's progress in https://circleci.com/gh/giantswarm/azure-admission-controller/.

## Edit the release draft and publish

Open the [release draft](https://github.com/giantswarm/azure-admission-controller/releases/) on Github.

Edit the description to inform about what has changed since the last release. Save and publish the release.

## Prerequisites

CircleCI must be set up with certain environment variables:

- `CODE_SIGNING_CERT_BUNDLE_BASE64` - Base64 encoded PKCS#12 key/cert bundle used for signing Windows binaries
- `CODE_SIGNING_CERT_BUNDLE_PASSWORD` - Password for the above bundle
- `RELEASE_TOKEN` - A GitHub token with the permission to write to repositories
  - [giantswarm/azure-admission-controller](https://github.com/giantswarm/azure-admission-controller/)
- `GITHUB_USER_EMAIL` - Email address of the github user owning the personal token above
- `GITHUB_USER_NAME` - Username of the above github user
