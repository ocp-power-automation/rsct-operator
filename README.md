# RSCT Operator

The `RSCT` Operator enables RSCT on a cluster.

- [Deploying the `RSCT` Operator](#deploying-the-rsct-operator)
    - [Installing the `RSCT` Operator](#installing-the-rsct-operator)
    - [Create `RSCT` resource](#create-rsct-resource)
    - [Check that the Operator is Running](#check-if-the-operator-is-running)
    - [Removing the `RSCT` Operator](#remove-the-rsct-operator)

### Installing the `RSCT` Operator
Run the following command to deploy the `RSCT` Operator:
```sh
make deploy
```

### Create `RSCT` resource
1. Switch to the `rsct-operator-system` namespace
   ```sh
   oc project rsct-operator-system
   ```

2. Create the cluster role binding and `RSCT` resources
   ```sh
   oc apply -f config/samples/rsct_v1alpha1_rsct.yaml
   ```

### Check that the Operator is Running
Check that the pods have been created and are running
```sh
oc get pods
```

A readiness probe will check that each pod is running, which can take up to 10 minutes.

Output should be similar to the following
```
NAME                                              READY   STATUS    RESTARTS   AGE
rsct-operator-controller-manager-b8495fb5-kznsw   2/2     Running   0          10m
rsct-test-9qstd                                   1/1     Running   0          6m49s
rsct-test-gclb2                                   1/1     Running   0          6m49s
rsct-test-gcqjh                                   1/1     Running   0          6m49s
rsct-test-tbvdx                                   1/1     Running   0          6m49s
rsct-test-wt8x9                                   1/1     Running   0          6m49s
```

### Removing the `RSCT` Operator
Run the following command to remove the `RSCT` Operator:
```sh
make undeploy
```
