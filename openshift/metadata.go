package openshift

import (
	fn "knative.dev/kn-plugin-func"
)

const (
	annotationOpenShiftVcsUri = "app.openshift.io/vcs-uri"
	annotationOpenShiftVcsRef = "app.openshift.io/vcs-ref"

	labelAppK8sInstance   = "app.kubernetes.io/instance"
	labelOpenShiftRuntime = "app.openshift.io/runtime"
)

var iconValuesForRuntimes = map[string]string{
	"go":         "golang",
	"node":       "nodejs",
	"python":     "python",
	"quarkus":    "quarkus",
	"springboot": "spring-boot",
}

var s2iTektonTaskNameForRuntime = map[string]string{
	"node":       "s2i-nodejs",
	"typescript": "s2i-nodejs",
	"quarkus":    "s2i-java",
}

type OpenshiftMetadataDecorator struct{}

func (o OpenshiftMetadataDecorator) UpdateAnnotations(f fn.Function, annotations map[string]string) map[string]string {
	if annotations == nil {
		annotations = map[string]string{}
	}
	if f.Git.URL != nil {
		annotations[annotationOpenShiftVcsUri] = *f.Git.URL
	}
	if f.Git.Revision != nil {
		annotations[annotationOpenShiftVcsRef] = *f.Git.Revision
	}

	return annotations
}

func (o OpenshiftMetadataDecorator) UpdateLabels(f fn.Function, labels map[string]string) map[string]string {
	if labels == nil {
		labels = map[string]string{}
	}

	// this label is used for referencing a Tekton Pipeline and deployed KService
	labels[labelAppK8sInstance] = f.Name

	// if supported, set the label representing a runtime icon in Developer Console
	iconValue, ok := iconValuesForRuntimes[f.Runtime]
	if ok {
		labels[labelOpenShiftRuntime] = iconValue
	}

	return labels
}

// GetS2iTektonTaskProperties returns a kind and name of Tekton Task then can be used on OpenShift for S2I builds
// Also returns whether the task should define Builder Image Parameter
func (o OpenshiftMetadataDecorator) GetS2iTektonTaskProperties(f fn.Function) (kind, taskName string, defineBuilderImageParam bool) {
	kind = "ClusterTask"

	// don't define Builder image param -> let's use the default builder image for each task on OpenShift
	defineBuilderImageParam = false

	// select the proper TektonTask
	name, ok := s2iTektonTaskNameForRuntime[f.Runtime]
	if ok {
		taskName = name
	}

	return
}
