package resource

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/alexandrevilain/sipp-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ConfigMapBuilder struct {
	Instance *v1alpha1.SippScenarioRun
	Scenario *v1alpha1.SippScenario
	Scheme   *runtime.Scheme
}

func NewConfigMapBuilder(builder *SippResourceBuilder) *ConfigMapBuilder {
	return &ConfigMapBuilder{
		Instance: builder.Instance,
		Scenario: builder.Scenario,
		Scheme:   builder.Scheme,
	}
}

func (b *ConfigMapBuilder) getLabelsSelectors() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      b.Instance.ChildResourceName("configmap"),
		"app.kubernetes.io/component": "configmap",
		"app.kubernetes.io/part-of":   "sipp-run",
	}
}

func (b *ConfigMapBuilder) getAnnotations() map[string]string {
	return map[string]string{}
}

func (b *ConfigMapBuilder) Build() (runtime.Object, error) {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.Instance.ChildResourceName("configmap"),
			Namespace:   b.Instance.Namespace,
			Labels:      b.getLabelsSelectors(),
			Annotations: b.getAnnotations(),
		},
	}, nil
}

func (b *ConfigMapBuilder) Update(object runtime.Object) error {
	configMap := object.(*corev1.ConfigMap)

	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}

	if b.Scenario.Spec.ScenarioFileContent != "" {
		configMap.Data["scenario.xml"] = b.Scenario.Spec.ScenarioFileContent
	}

	for i, value := range b.Scenario.Spec.InjectValues {
		configMap.Data[b.Scenario.GetInjectedValueFilename(i)] = value
	}

	if err := controllerutil.SetControllerReference(b.Instance, configMap, b.Scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}
