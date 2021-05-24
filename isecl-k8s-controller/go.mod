module intel/isecl/k8s-custom-controller/v4

require (
	github.com/intel-secl/intel-secl/v4 v4.0.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.5-beta.0
	k8s.io/client-go v0.17.4
)

replace github.com/intel-secl/intel-secl/v4 => gitlab.devtools.intel.com/sst/isecl/intel-secl.git/v4 v4.0/develop
