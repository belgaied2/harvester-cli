module github.com/belgaied2/harvester-cli

go 1.16

replace (
	github.com/openshift/api => github.com/openshift/api v0.0.0-20191219222812-2987a591a72c
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20210112165513-ebc401615f47
	github.com/operator-framework/operator-registry => github.com/operator-framework/operator-registry v1.17.4
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.2
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v0.0.0-20190302045857-e85c7b244fd2
)

require (
	github.com/docker/docker v20.10.7+incompatible // indirect
	github.com/go-kit/kit v0.11.0 // indirect
	github.com/hashicorp/go-version v1.3.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rancher/cli v1.0.0-alpha9.0.20210315153654-8de9f8e29aef
	github.com/rancher/norman v0.0.0-20210608202517-59b3523c3133 // indirect
	github.com/rancher/types v0.0.0-20200528213132-b5fb46b1825d
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli v1.22.5
	k8s.io/apimachinery v0.20.6
	kubevirt.io/client-go v0.41.0
)
