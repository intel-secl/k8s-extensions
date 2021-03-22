module intel/isecl/k8s-extended-scheduler/v3

require (
	github.com/Waterdrips/jwt-go v3.2.1-0.20200915121943-f6506928b72e+incompatible
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/intel-secl/intel-secl/v3 v3.5.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	k8s.io/api v0.18.2
	k8s.io/kube-scheduler v0.18.2

)

replace github.com/intel-secl/intel-secl/v3 => github.com/intel-secl/intel-secl/v3 v3.5.0
