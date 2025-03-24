# Building the RSCT Images
### Export Environment Variables
Before building the images, export the necessary environment variables if reuired.  
For example:
```
# Registry where images will be pushed
# Default value is ghcr.io/ocp-power-automation/rsct-operator
export IMAGE_TAG_BASE=ghcr.io/ocp-power-automation/rsct-operator

# Current version of the operator to build
# Default value is set to the current released version at the ghcr.io/ocp-power-automation/rsct-operator registry
export VERSION=0.0.1-alpha3

# Set the pervious version to the latest released operator version
# Default value is set to the pervious released version at the ghcr.io/ocp-power-automation/rsct-operator registry
export PREVIOUS_VERSION=0.0.1-alpha2

# Container tool (docker or podman)
# Default is set to podman
export CONTAINER_TOOL=docker

# To use a different registry for adding multiple operators in the new catalog source index, export the new registry and set the correct PREVIOUS_VERSION
# Default value is ghcr.io/ocp-power-automation/rsct-operator
export PREVIOUS_IMAGE_TAG_BASE=ghcr.io/ocp-power-automation/rsct-operator
```

### Update the Catalog Index
Before building the catalog, update the index `entries` in *catalog/index.yaml*. Ensure all `entries` match the `bundles` available in the catalog source.

For example, if you are building the following version:
```
export VERSION=0.0.1-alpha3
export PREVIOUS_VERSION=0.0.1-alpha2
```
Add the new entry for `VERSION=0.0.1-alpha3` in *catalog/index.yaml*:
```
  - name: rsct-operator.v0.0.1-alpha3
    replaces: rsct-operator.v0.0.1-alpha2
```
After adding the new version, the `entries` in [catalog/index.yaml](https://github.com/ocp-power-automation/rsct-operator/blob/main/catalog/index.yaml) will look like this:
```
---
schema: olm.channel
package: rsct-operator
name: alpha
entries:
  - name: rsct-operator.v0.0.1-alpha0
  - name: rsct-operator.v0.0.1-alpha1
    replaces: rsct-operator.v0.0.1-alpha0
  - name: rsct-operator.v0.0.1-alpha2
    replaces: rsct-operator.v0.0.1-alpha1
  - name: rsct-operator.v0.0.1-alpha3
    replaces: rsct-operator.v0.0.1-alpha2
```
**Note:**
Before pushing the images, make sure you are logged in to the registry:
```
docker login <registry>
```

## Building and Pushing Images
1. Build and push the operator image

**Single architecture image (ppc64le)**  
```
make docker-build docker-push
```

**Multi-architecture image**  
```
make docker-buildx
```

2. Build and push the bundle image
```
make bundle bundle-build bundle-push
```

3. Build and push the catalog images
```
make catalog-build catalog-push
````
This will build and push the catalog image with the version tag `v0.0.1-alpha3` and `latest`.