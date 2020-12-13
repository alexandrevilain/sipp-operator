package resource

import (
	"github.com/alexandrevilain/sipp-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ResourceBuilder interface {
	Build() (runtime.Object, error)
	Update(runtime.Object) error
}

type SippResourceBuilder struct {
	Instance *v1alpha1.SippScenarioRun
	Scenario *v1alpha1.SippScenario
	Scheme   *runtime.Scheme
}

func (builder *SippResourceBuilder) ResourceBuilders() ([]ResourceBuilder, error) {
	return []ResourceBuilder{
		NewConfigMapBuilder(builder),
		NewJobBuilder(builder),
	}, nil
}
