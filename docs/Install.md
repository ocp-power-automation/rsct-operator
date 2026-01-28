## RSCT Deployment on OpenShift Clusters

### Prerequisites for OpenShift Clusters
- OpenShift cluster with at least one Power node and admin access.
- Create a CatalogSource using the image `ghcr.io/ocp-power-automation/rsct-operator-catalog:latest` in the `openshift-marketplace` namespace.
```
oc create -f - <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: rsct-catalogsource
  namespace: openshift-marketplace
spec:
  displayName: RSCT Operator
  publisher: IBM
  sourceType: grpc
  image: ghcr.io/ocp-power-automation/rsct-operator-catalog:latest
EOF
```
- Use `rsct-operator-system` namespace for RSCT deployment.
    
### Operator Deployment
1. Create the `rsct-operator-system` namespace
```
oc create -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: rsct-operator-system
EOF
```

2. Create an OperatorGroup for the RSCT Operator
```
oc create -f - <<EOF
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: rsct-operator-operatorgroup
  namespace: rsct-operator-system
spec:
  targetNamespaces:
  - rsct-operator-system
EOF
```

3. Create an Subscription for the RSCT Operator
```
oc create -f - <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: rsct-operator-subscription
  namespace: rsct-operator-system
spec:
  channel: "alpha"
  installPlanApproval: Automatic
  name: rsct-operator
  source: rsct-catalogsource
  sourceNamespace: openshift-marketplace
EOF
```

Note:
To install specific operator version
  - Use the `startingCSV` field within the spec and set `installPlanApproval` to `Manual`. 

Example:
```
oc create -f - <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: rsct-operator-subscription
  namespace: rsct-operator-system
spec:
  channel: "alpha"
  installPlanApproval: Manual
  name: rsct-operator
  source: rsct-catalogsource
  sourceNamespace: openshift-marketplace
  startingCSV: rsct-operator.v0.0.1-alpha0
EOF
```
  - Approve the installation of the operator by updating the approved field of the InstallPlan:
```
oc patch installplan install-abcd \
    --namespace rsct-operator-system \
    --type merge \
    --patch '{"spec":{"approved":true}}'
```
4. Verify the operator deployment
```
# oc get csv -n rsct-operator-system
NAME                           DISPLAY                                      VERSION        REPLACES                      PHASE
rsct-operator.v0.0.1-alpha3    RSCT Operator for IBM Power Systems          0.0.1-alpha3                                 Succeeded

# oc get pods -n rsct-operator-system
NAME                                                READY   STATUS    RESTARTS   AGE
rsct-operator-controller-manager-7978b6f4cd-kjsmd   2/2     Running   0          28s
```

###  Create the Custom Resource
After deploying the operator, you must create an RSCT object. The following example will create an RSCT instance.
```
oc create -f - <<EOF
apiVersion: rsct.ibm.com/v1alpha1
kind: RSCT
metadata:
  name: rsct
  namespace: rsct-operator-system
spec: {}
EOF
```

The RSCT image can be configured in the spec.
```
spec:
  image: quay.io/powercloud/rsct-ppc64le:latest
```
Default image is `quay.io/powercloud/rsct-ppc64le:latest`

**Verify the RSCT deployment:**
```
# oc get pods -n rsct-operator-system
NAME                                                READY   STATUS    RESTARTS   AGE
rsct-brs49                                          1/1     Running   0          13s
rsct-c2q54                                          1/1     Running   0          13s
rsct-mvfm4                                          1/1     Running   0          13s
rsct-mwbn6                                          1/1     Running   0          13s
rsct-operator-controller-manager-7978b6f4cd-kjsmd   2/2     Running   0          66s
rsct-r6tdl                                          1/1     Running   0          13s
```
