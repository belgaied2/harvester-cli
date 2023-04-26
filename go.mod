module github.com/belgaied2/harvester-cli

go 1.16

replace (
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go v3.2.1-0.20200107013213-dc14462fd587+incompatible
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/docker/docker v1.4.2-0.20200203170920-46ec8731fbce
	github.com/go-kit/kit => github.com/go-kit/kit v0.3.0
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1
	github.com/knative/pkg => github.com/rancher/pkg v0.0.0-20190514055449-b30ab9de040e
	github.com/openshift/api => github.com/openshift/api v0.0.0-20191219222812-2987a591a72c
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20200521150516-05eb9880269c
	github.com/operator-framework/operator-lifecycle-manager => github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190128024246-5eb7ae5bdb7a
	github.com/rancher/rancher/pkg/apis => github.com/rancher/rancher/pkg/apis v0.0.0-20211208233239-77392a65423d
	github.com/rancher/rancher/pkg/client => github.com/rancher/rancher/pkg/client v0.0.0-20211208233239-77392a65423d

	helm.sh/helm/v3 => github.com/rancher/helm/v3 v3.5.4-rancher.1
	k8s.io/api => k8s.io/api v0.24.7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.24.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.24.7
	k8s.io/apiserver => k8s.io/apiserver v0.24.7
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.24.7
	k8s.io/client-go => k8s.io/client-go v0.24.7
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.24.7
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.24.7
	k8s.io/code-generator => k8s.io/code-generator v0.24.7
	k8s.io/component-base => k8s.io/component-base v0.24.7
	k8s.io/component-helpers => k8s.io/component-helpers v0.24.7
	k8s.io/controller-manager => k8s.io/controller-manager v0.24.7
	k8s.io/cri-api => k8s.io/cri-api v0.24.7
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.24.7
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.24.7
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.24.7
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.24.7
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.24.7
	k8s.io/kubectl => k8s.io/kubectl v0.24.7
	k8s.io/kubelet => k8s.io/kubelet v0.24.7
	k8s.io/kubernetes => k8s.io/kubernetes v1.23.7
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.24.7
	k8s.io/metrics => k8s.io/metrics v0.24.7
	k8s.io/mount-utils => k8s.io/mount-utils v0.24.7
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.24.7
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.24.7
	kubevirt.io/api => github.com/kubevirt/api v0.54.0

	kubevirt.io/client-go => github.com/kubevirt/client-go v0.54.0
	kubevirt.io/containerized-data-importer => github.com/rancher/kubevirt-containerized-data-importer v1.26.1-0.20210802100720-9bcf4e7ba0ce
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v0.4.4
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v0.0.0-20190302045857-e85c7b244fd2
)

require (
	github.com/docker/docker v20.10.12+incompatible
	github.com/google/gxui v0.0.0-20151028112939-f85e0a97b3a4 // indirect
	github.com/grantae/certinfo v0.0.0-20170412194111-59d56a35515b
	github.com/harvester/harvester v1.1.1
	github.com/harvester/vm-import-controller v0.1.4 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/minio/pkg v1.1.14
	github.com/pkg/errors v0.9.1
	github.com/rancher/cli v1.0.0-alpha9.0.20210315153654-8de9f8e29aef
	github.com/rancher/norman v0.0.0-20220520225714-4cc2f5a97011
	github.com/rancher/types v0.0.0-20210123000350-7cb436b3f0b0
	github.com/rancher/wrangler v1.1.0 // indirect
	github.com/sirupsen/logrus v1.9.0
	github.com/urfave/cli v1.22.5
	github.com/urfave/cli/v2 v2.25.1
	github.com/zach-klippenstein/goregen v0.0.0-20160303162051-795b5e3961ea
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4 // indirect
	golang.org/x/term v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools/v3 v3.2.0 // indirect
	k8s.io/api v0.25.4
	k8s.io/apimachinery v0.25.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kubectl v0.24.7
	kubevirt.io/api v0.59.0
	kubevirt.io/client-go v0.49.0 // indirect
	kubevirt.io/containerized-data-importer-api v1.50.0 // indirect
	sigs.k8s.io/controller-runtime v0.12.2 // indirect
)
