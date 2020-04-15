module github.com/danieloliveira079/howto-agones-informers

go 1.14

require (
	agones.dev/agones v1.4.0
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/crypto v0.0.0-20200414173820-0848c9571904 // indirect
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	// Agones supports Kubernetes 1.14. Make sure you are using the right K8S packages
	k8s.io/api v0.0.0-20191004102349-159aefb8556b // kubernetes-1.14.10
	k8s.io/apiextensions-apiserver v0.0.0-20191212015246-8fe0c124fb40 // kubernetes-1.14.10
	k8s.io/apimachinery v0.0.0-20191004074956-c5d2f014d689 // kubernetes-1.14.10
	k8s.io/client-go v11.0.1-0.20191029005444-8e4128053008+incompatible // kubernetes-1.14.10
	k8s.io/klog v1.0.0 // indirect
	k8s.io/utils v0.0.0-20200414100711-2df71ebbae66 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
