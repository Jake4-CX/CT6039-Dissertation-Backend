# Kubernetes Deployment for Load Testing Application

This document outlines the steps to deploy the Load Testing application to a Kubernetes cluster using Minikube. It includes instructions for deploying the application components, as well as accessing the REST API exposed by the `loadtest-master` service.

> [!NOTE]
> TLS is enabled by default, requring the application to run on port 443 and is only accessible on the defined domain.

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/) installed and configured to communicate with your Kubernetes cluster.
- [Minikube](https://minikube.sigs.k8s.io/docs/start/) installed for local Kubernetes deployment. Ensure Minikube is started using `minikube start`.

## Deployment

To deploy the Load Testing application to your Kubernetes cluster, follow these steps:

1. **Navigate to the Deployment Directory**: Change your current working directory to the `k8s` directory inside `deployments`.

    ```bash
    cd /deployments/k8s/
    ```

2. **Apply the Kubernetes Configurations**: Deploy the application components (ConfigMap, PersistentVolumeClaim, Deployments, and Services) to your Kubernetes cluster.

    ```bash
    kubectl apply -f ./
    ```

    This command will create all the necessary Kubernetes resources defined in the `.yaml` files within the directory.

## Accessing the Application

After deploying the application, you can access the REST API exposed by the `loadtest-master` service using Minikube.

### Setup Minikube Tunnel

Minikube tunnel creates a route on your local machine that maps the LoadBalancer service type to an accessible IP address. Open a new terminal window and run:

```bash
minikube tunnel
```

**Note**: Keep the terminal running with the tunnel command for as long as you need access to the application.

## Access the REST API

With the tunnel running, use the following command to find the IP address and port assigned to the loadtest-master service:

```bash
kubectl get service loadtest-master
```

Look for the `EXTERNAL-IP` and `PORT(S)` values. You can now access the REST API using the provided IP address and port:

```bash
http://<EXTERNAL-IP>:<PORT>
```

Replace `<EXTERNAL-IP>` and `<PORT>` with the actual values obtained from the previous command.

## Cleanup

To remove the deployed resources from your cluster, run:

```bash
kubectl delete -f ./
```

This will delete the Kubernetes resources created for this application.

## Restart

To restart the deployed resources, run:

```bash
kubectl rollout restart deployment <DEPLOYMENT_NAME>
```

Replace `<DEPLOYMENT_NAME>` with the name of the container, i.e. `loadtest-worker` or `loadtest-master`. Node this restart does not introduce downtime.
