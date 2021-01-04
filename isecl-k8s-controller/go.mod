module intel/isecl/k8s-custom-controller/v3

require (
	github.com/intel-secl/intel-secl/v3 v3.3.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	k8s.io/api v0.17.4
	k8s.io/apiextensions-apiserver v0.17.4
	k8s.io/apimachinery v0.17.5-beta.0
	k8s.io/client-go v0.17.4
)

replace github.com/intel-secl/intel-secl/v3 => gitlab.devtools.intel.com/sst/isecl/intel-secl.git/v3 v3.3.1/develop

