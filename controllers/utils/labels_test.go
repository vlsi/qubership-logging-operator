package utils

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetLabelsForWorkload_ComponentLabelsOnPodTemplate(t *testing.T) {
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ds",
			Labels: map[string]string{
				"existing": "label",
			},
		},
	}
	ds.Spec.Template.Labels = map[string]string{"pod-existing": "value"}

	in := LabelInput{
		Name:      "test-ds",
		Component: "fluentbit",
		Instance:  "test-ds-ns",
		Version:   "1.0",
		Technology: "c",
		ComponentLabels: map[string]string{
			"custom-label":  "custom-value",
			"another-label": "another-value",
		},
	}

	SetLabelsForWorkload(ds, &ds.Spec.Template.Labels, in)

	// Verify ComponentLabels are present on the resource
	for k, v := range in.ComponentLabels {
		if got := ds.Labels[k]; got != v {
			t.Errorf("resource label %q = %q, want %q", k, got, v)
		}
	}

	// Verify ComponentLabels are present on the pod template
	for k, v := range in.ComponentLabels {
		if got := ds.Spec.Template.Labels[k]; got != v {
			t.Errorf("pod template label %q = %q, want %q", k, got, v)
		}
	}

	// Verify base labels are still present on the pod template
	wantBase := map[string]string{
		"name":                        "test-ds",
		"app.kubernetes.io/name":      "test-ds",
		"app.kubernetes.io/component": "fluentbit",
		"app.kubernetes.io/part-of":   PartOfLogging,
		"app.kubernetes.io/managed-by": ManagedByOperator,
		"app.kubernetes.io/managed-by-operator": OperatorDeploymentName,
		"app.kubernetes.io/instance":   "test-ds-ns",
		"app.kubernetes.io/version":    "1.0",
		"app.kubernetes.io/technology": "c",
	}
	for k, v := range wantBase {
		if got := ds.Spec.Template.Labels[k]; got != v {
			t.Errorf("pod template base label %q = %q, want %q", k, got, v)
		}
	}
}

func TestSetLabelsForWorkload_NilComponentLabels(t *testing.T) {
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-ds",
			Labels: map[string]string{},
		},
	}
	ds.Spec.Template.Labels = map[string]string{}

	in := LabelInput{
		Name:      "test-ds",
		Component: "fluentbit",
		Instance:  "test-ds-ns",
		Version:   "1.0",
	}

	SetLabelsForWorkload(ds, &ds.Spec.Template.Labels, in)

	// Should not panic and should have base labels
	if got := ds.Spec.Template.Labels["app.kubernetes.io/name"]; got != "test-ds" {
		t.Errorf("pod template label app.kubernetes.io/name = %q, want %q", got, "test-ds")
	}
}

func TestSetLabelsForWorkload_ComponentLabelsOverrideBase(t *testing.T) {
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-ds",
			Labels: map[string]string{},
		},
	}
	ds.Spec.Template.Labels = map[string]string{}

	in := LabelInput{
		Name:      "test-ds",
		Component: "fluentbit",
		ComponentLabels: map[string]string{
			"app.kubernetes.io/component": "custom-component",
		},
	}

	SetLabelsForWorkload(ds, &ds.Spec.Template.Labels, in)

	// ComponentLabels should override base labels on both resource and pod template
	if got := ds.Labels["app.kubernetes.io/component"]; got != "custom-component" {
		t.Errorf("resource label app.kubernetes.io/component = %q, want %q", got, "custom-component")
	}
	if got := ds.Spec.Template.Labels["app.kubernetes.io/component"]; got != "custom-component" {
		t.Errorf("pod template label app.kubernetes.io/component = %q, want %q", got, "custom-component")
	}
}
