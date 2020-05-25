# What is it?

Dynamic version of ExtendedResourceToleration plugin for Kubernetes.

It automatically adds tolerations if Pod definition has extended resource requests/limits, like:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: gpu-app
spec:
  containers:
    - name: gpu-app
      image: "some-gpu-required-image:latest"
      resources:
        limits:
          nvidia.com/gpu: 1 # requesting 1 GPU
```

into

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: gpu-app
spec:
  containers:
    - name: gpu-app
      image: "some-gpu-required-image:latest"
      resources:
        limits:
          nvidia.com/gpu: 1 # requesting 1 GPU
  tolerations:
    - key: "nvidia.com/gpu"
      operator: "Exists"
      effect: "NoSchedule"
```

dynamically when scheduling to create a pod.

See official [Documentation](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#extendedresourcetoleration) for detail.

# Usage

You need some dependencies:

- GNU make
- [kustomize](https://kustomize.io/)
- openssl

```bash
git clone https://github.com/roy-ht/extended-resource-toleration.git
cd extended-resource-toleration
# You can check manifests if needed
KS_NAMESPACE=default KS_ARG="--dryrun" make apply-k8s
KS_NAMESPACE=default make apply-k8s
```

# Acknowledgements

Some part of codes are derived from below:

- [kubeflow/admission-webhook](https://github.com/kubeflow/kubeflow/tree/master/components/admission-webhook)
- [kubernetes/plugin/extendedresourcetoleration](https://github.com/kubernetes/kubernetes/tree/master/plugin/pkg/admission/extendedresourcetoleration)
