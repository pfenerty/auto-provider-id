kubectl create ns skaffold-builds
kubectl label ns skaffold-builds pod-security.kubernetes.io/enforce=privileged
kubectl label ns skaffold-builds pod-security.kubernetes.io/warn=privileged
kubectl -n skaffold-builds create secret generic build-secret --from-file=config.json=/home/patrick/.docker/config.json

kubectl create ns auto-provider-id