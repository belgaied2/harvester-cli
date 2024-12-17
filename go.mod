module github.com/belgaied2/harvester-cli

go 1.23.3

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
	github.com/grantae/certinfo v0.0.0-20170412194111-59d56a35515b
	github.com/harvester/harvester v1.1.1
	github.com/harvester/vm-import-controller v0.1.4
	github.com/minio/pkg v1.1.14
	github.com/pkg/errors v0.9.1
	github.com/rancher/cli v1.0.0-alpha9.0.20210315153654-8de9f8e29aef
	github.com/rancher/norman v0.0.0-20220520225714-4cc2f5a97011
	github.com/rancher/types v0.0.0-20210123000350-7cb436b3f0b0
	github.com/sirupsen/logrus v1.9.0
	github.com/urfave/cli v1.22.5
	github.com/urfave/cli/v2 v2.25.1
	github.com/zach-klippenstein/goregen v0.0.0-20160303162051-795b5e3961ea
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.25.4
	k8s.io/apimachinery v0.25.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kubectl v0.24.7
	kubevirt.io/api v0.59.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/c-bata/go-prompt v0.2.6 // indirect
	github.com/c9s/goprocinfo v0.0.0-20210130143923-c95fcf8c64a8 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/chai2010/gettext-go v0.0.0-20160711120539-c6fed771bfd5 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.8.0 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/felixge/httpsnoop v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/fvbommel/sortorder v1.0.1 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/gxui v0.0.0-20151028112939-f85e0a97b3a4 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/go-version v1.2.1 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jinzhu/copier v0.3.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/k8snetworkplumbingwg/network-attachment-definition-client v0.0.0-20200331171230-d50e42f2b669 // indirect
	github.com/kubernetes-csi/external-snapshotter/v2 v2.1.1 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/longhorn/go-iscsi-helper v0.0.0-20220805034259-7b59e22574bb // indirect
	github.com/longhorn/longhorn-manager v1.3.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mattn/go-tty v0.0.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/openshift/custom-resource-status v1.1.2 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/term v1.2.0-beta.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.52.0 // indirect
	github.com/prometheus/client_golang v1.12.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rancher/lasso v0.0.0-20221227210133-6ea88ca2fbcc // indirect
	github.com/rancher/system-upgrade-controller/pkg/apis v0.0.0-20210727200656-10b094e30007 // indirect
	github.com/rancher/wrangler v1.1.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday v1.5.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spf13/cobra v1.4.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.8.1 // indirect
	github.com/xlab/treeprint v1.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.starlark.net v0.0.0-20200306205701-8dd3e2ee1dd5 // indirect
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4 // indirect
	golang.org/x/net v0.0.0-20220722155237-a158d28d115b // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/term v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apiextensions-apiserver v0.25.4 // indirect
	k8s.io/cli-runtime v0.24.7 // indirect
	k8s.io/component-base v0.24.7 // indirect
	k8s.io/klog/v2 v2.70.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220803162953-67bda5d908f1 // indirect
	k8s.io/utils v0.0.0-20220728103510-ee6ede2d64ed // indirect
	kubevirt.io/containerized-data-importer-api v1.50.0 // indirect
	kubevirt.io/controller-lifecycle-operator-sdk/api v0.0.0-20220329064328-f3cc58c6ed90 // indirect
	kubevirt.io/kubevirt v0.54.0 // indirect
	sigs.k8s.io/cluster-api v0.4.4 // indirect
	sigs.k8s.io/controller-runtime v0.12.2 // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/kustomize/api v0.11.4 // indirect
	sigs.k8s.io/kustomize/kyaml v0.13.6 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
