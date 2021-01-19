module github.com/jenkins-x-plugins/jx-tekton-to-actions

require (
	github.com/aws/aws-sdk-go v1.35.18 // indirect
	github.com/cpuguy83/go-md2man v1.0.10
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/jenkins-x/go-scm v1.5.215 // indirect
	github.com/jenkins-x/jx-api/v4 v4.0.23 // indirect
	github.com/jenkins-x/jx-gitops v0.0.531 // indirect
	github.com/jenkins-x/jx-helpers/v3 v3.0.65
	github.com/jenkins-x/jx-logging/v3 v3.0.3
		github.com/jenkins-x/lighthouse-client v0.0.0-20210118141307-27a29c02a663
	github.com/kr/pretty v0.2.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/tektoncd/pipeline v0.20.0
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0 // indirect
)

replace (
	github.com/tektoncd/pipeline => github.com/jenkins-x/pipeline v0.3.2-0.20210118090417-1e821d85abf6
	k8s.io/client-go => k8s.io/client-go v0.20.2
	knative.dev/pkg => github.com/jstrachan/pkg v0.0.0-20210118084935-c7bdd6c14bd0
)
go 1.15
