# Integration tests

```bash
(kind delete cluster || true) && kind create cluster && apptestctl bootstrap --kubeconfig="$(kind get kubeconfig)"
E2E_KUBECONFIG=~/.kube/config CIRCLE_SHA1=$(git rev-parse HEAD) AZURE_CLIENTID="${AZURE_CLIENTID}" AZURE_CLIENTSECRET="${AZURE_CLIENTSECRET}" AZURE_TENANTID="${AZURE_TENANTID}" AZURE_SUBSCRIPTIONID="${AZURE_SUBSCRIPTIONID}" go test -tags=k8srequired ./integration/test/createcluster -count=1 | luigi
```
