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
oc create -k "https://github.com/ocp-power-automation/rsct-operator/config/default/?ref=main"
```

### Create `RSCT` resource
Run the following command to deploy the `RSCT` resource
```sh
oc create -k "https://github.com/ocp-power-automation/rsct-operator/config/samples/?ref=main"
```

### Check that the Operator is Running
Check that the pods have been created and are running
```sh
oc get pods -n rsct-operator-system
```

A readiness probe will check that each pod is running, which can take up to 10 minutes.

Output should be similar to the following
```
NAME                                                READY   STATUS    RESTARTS   AGE
rsct-6tbjw                                          1/1     Running   0          7s
rsct-fr8tv                                          1/1     Running   0          7s
rsct-h4lts                                          1/1     Running   0          7s
rsct-mxwgt                                          1/1     Running   0          7s
rsct-operator-controller-manager-5d4c458d86-xln94   2/2     Running   0          62s
rsct-rgvlm                                          1/1     Running   0          7s
```

### Removing the `RSCT` Operator
Run the following command to remove the `RSCT` Operator:
1. Remove the cluster role binding and `RSCT` resources
```sh
oc delete -k "https://github.com/ocp-power-automation/rsct-operator/config/samples/?ref=main"
```
2. Remove the `RSCT` Operator:
```
oc delete -k "https://github.com/ocp-power-automation/rsct-operator/config/default/?ref=main"
```
