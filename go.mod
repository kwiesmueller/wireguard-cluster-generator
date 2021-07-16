module github.com/kwiesmueller/wireguard-cluster-gen

go 1.16

replace (
	k8s.io/api => github.com/rancher/kubernetes/staging/src/k8s.io/api v1.16.3-k3s.2
	k8s.io/apiextensions-apiserver => github.com/rancher/kubernetes/staging/src/k8s.io/apiextensions-apiserver v1.16.3-k3s.2
	k8s.io/apimachinery => github.com/rancher/kubernetes/staging/src/k8s.io/apimachinery v1.16.3-k3s.2
	k8s.io/apiserver => github.com/rancher/kubernetes/staging/src/k8s.io/apiserver v1.16.3-k3s.2
	k8s.io/client-go => github.com/rancher/kubernetes/staging/src/k8s.io/client-go v1.16.3-k3s.2
	k8s.io/code-generator => github.com/rancher/kubernetes/staging/src/k8s.io/code-generator v1.16.3-k3s.2
	k8s.io/component-base => github.com/rancher/kubernetes/staging/src/k8s.io/component-base v1.16.3-k3s.2
	k8s.io/kube-aggregator => github.com/rancher/kubernetes/staging/src/k8s.io/kube-aggregator v1.16.3-k3s.2
	k8s.io/metrics => github.com/rancher/kubernetes/staging/src/k8s.io/metrics v1.16.3-k3s.2
)

require (
	github.com/rancher/k3os v0.11.1
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20210506160403-92e472f520a5
	gopkg.in/yaml.v2 v2.2.4
)
