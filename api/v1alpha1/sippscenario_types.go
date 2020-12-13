/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SippScenarioSpec defines the desired state of SippScenario
type SippScenarioSpec struct {
	// ScenarioFileContent
	// See the -sf parameter documentation
	ScenarioFileContent string `json:"scenarioFileContent"`
	// InjectValues is the file content which allow to values from an external CSV file during calls into the scenarios.
	// See the -inf parameter documentation
	// +optional
	InjectValues []string `json:"injectValues"`
}

// +kubebuilder:object:root=true

// SippScenario is the Schema for the sippscenarios API
// +kubebuilder:resource:shortName={"ss"}
type SippScenario struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SippScenarioSpec `json:"spec,omitempty"`
}

// GetInjectedValueFilename returns the filename of the inject values file
// from its id
func (f *SippScenario) GetInjectedValueFilename(i int) string {
	return fmt.Sprintf("values_%d.csv", i)
}

// ToSippArgs returns the Sipp Args from the Spec
// This function asserts that the Spec is clean (no unknown values)
func (f *SippScenario) ToSippArgs(basePath string) []string {
	result := []string{}

	if f.Spec.ScenarioFileContent != "" {
		result = append(result, "-sf")
		result = append(result, fmt.Sprintf("%s/scenario.xml", basePath))
	}

	if len(f.Spec.InjectValues) > 0 {
		for i := range f.Spec.InjectValues {
			path := fmt.Sprintf("%s/%s", basePath, f.GetInjectedValueFilename(i))

			result = append(result, "-inf")
			result = append(result, path)
		}
	}

	return result
}

// +kubebuilder:object:root=true

// SippScenarioList contains a list of SippScenario
type SippScenarioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SippScenario `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SippScenario{}, &SippScenarioList{})
}
