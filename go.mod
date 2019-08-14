module github.com/aflc/extended-resource-toleration-webhook

go 1.12

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628
)

require (
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/kubeflow/kubeflow/components/admission-webhook v0.0.0-20190814035625-b376b38268ce // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	k8s.io/api v0.0.0-20190813220812-4c9d9526570f
	k8s.io/apiextensions-apiserver v0.0.0-20190814022005-b065192e8893 // indirect
	k8s.io/apimachinery v0.0.0-20190813235223-d2c4b5819cd0
	k8s.io/klog v0.4.0
	sigs.k8s.io/controller-runtime v0.1.12
)
