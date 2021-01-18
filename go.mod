module github.com/jenkins-x-plugins/jx-tekton-to-actions

require (
	cloud.google.com/go v0.70.0 // indirect
	github.com/aws/aws-sdk-go v1.35.18 // indirect
	github.com/cpuguy83/go-md2man v1.0.10
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/jenkins-x/go-scm v1.5.211 // indirect
	github.com/jenkins-x/jx-api/v4 v4.0.23 // indirect
	github.com/jenkins-x/jx-gitops v0.0.528 // indirect
	github.com/jenkins-x/jx-helpers/v3 v3.0.63
	github.com/jenkins-x/jx-logging/v3 v3.0.3
	github.com/jenkins-x/lighthouse v0.0.908
	github.com/kr/pretty v0.2.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/tektoncd/pipeline v0.16.3
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9 // indirect
)

replace (
	github.com/jenkins-x/lighthouse => github.com/jstrachan/lighthouse v0.0.0-20201116155709-614d66231eb3
	github.com/tektoncd/pipeline => github.com/jenkins-x/pipeline v0.0.0-20201002150609-ca0741e5d19a
	k8s.io/client-go => k8s.io/client-go v0.19.2
)

go 1.15
