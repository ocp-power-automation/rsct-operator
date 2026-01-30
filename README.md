# RSCT Operator

The `RSCT` Operator enables RSCT on a cluster.

- [Deploying the `RSCT` Operator](#deploying-the-rsct-operator)
    - [Installing the `RSCT` Operator](#installing-the-rsct-operator)
    - [Create `RSCT` resource](#create-rsct-resource)
    - [Check that the Operator is Running](#check-if-the-operator-is-running)
    - [Removing the `RSCT` Operator](#remove-the-rsct-operator)

### Installing the `RSCT` Operator

#### Option 1: Installing via CLI

1. Get the latest version of the RSCT Operator and set it as an environment variable:
```sh
export RSCT_VERSION=$(curl -s "https://api.github.com/repos/ocp-power-automation/rsct-operator/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
```

2. Run the following command to deploy the `RSCT` Operator using the version variable:
```sh
oc create -k "https://github.com/ocp-power-automation/rsct-operator/config/default/?ref=${RSCT_VERSION}"
```

#### Option 2: Installing via OperatorHub

You can also install the RSCT Operator directly from the OpenShift OperatorHub.

<img width="1470" height="878" alt="rsct" src="https://github.com/user-attachments/assets/01d359f1-8bc5-46f7-ab3d-3979b4f59a66" />


1. Navigate to **Operators** -> **OperatorHub** in the OpenShift Console.
2. Search for "RSCT Operator".
3. Click **Install**.

### Create `RSCT` resource
Run the following command to deploy the `RSCT` resource
```sh
oc create -k "https://github.com/ocp-power-automation/rsct-operator/config/samples/?ref=${RSCT_VERSION}"
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

#### Step 1: Remove the `RSCT` Resource

Remove the `RSCT` resource:
```sh
oc delete -k "https://github.com/ocp-power-automation/rsct-operator/config/samples/?ref=${RSCT_VERSION}"
```

#### Step 2: Remove the `RSCT` Operator

You can remove the operator using either the CLI or the OperatorHub.

**Option 1: Removing via CLI**

Run the following command to remove the `RSCT` Operator:
```sh
oc delete -k "https://github.com/ocp-power-automation/rsct-operator/config/default/?ref=${RSCT_VERSION}"
```

**Option 2: Removing via OperatorHub**

1. Navigate to **Operators** -> **Installed Operators** in the OpenShift Console.
2. Locate "RSCT Operator" in the list.
3. Click the options menu (three vertical dots) on the right side of the Operator entry.
4. Select **Uninstall Operator**.
5. When prompted, select **Uninstall** to remove the Operator, the Operator deployment, and pods.

