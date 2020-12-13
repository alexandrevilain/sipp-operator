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
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Protocol defines the protocol used in the scenario run
type Protocol string

const (
	// ProtocolTCP is the TCP protocol
	ProtocolTCP Protocol = "TCP"
	// ProtocolUDP is the UDP protocol
	ProtocolUDP Protocol = "UDP"
	// ProtocolTLS is the TLS protocol
	ProtocolTLS Protocol = "TLS"
)

// Socket defines the socket configuration of the scenario run
type Socket string

const (
	// SocketOne is one socket for all the scenario run
	SocketOne = "One"
	// SocketOnePerCall is one socket for each call of the scenario run
	SocketOnePerCall = "OnePerCall"
	// SocketOnePerIP is one socket for each ip of the scenario run
	SocketOnePerIP = "OnePerIP"
)

// Transport defines the transport used for the scenario run
type Transport struct {
	Protocol Protocol `json:"protocol"`
	Socket   Socket   `json:"socket"`
	// +optional
	Compression *bool `json:"compression"`
}

// SippScenarioRunSpec defines the desired state of SippScenarioRun
// TODO(alexandrevilain): Implement a Validating Admission Webhook
// If the CommandOverride is empty and the destination too, we should throw an error
type SippScenarioRunSpec struct {
	// ParallelismsSpecifies the maximum desired number of sipp instance you want to run at the same time
	// +optional
	Parallelism *int32 `json:"parallelism,omitempty"`
	// Sipp docker image
	// Defaults to ctaloi/sipp
	// +optional
	Image string `json:"image,omitempty"`
	// ImagePullSecrets is an optional list of references to secrets in the same namespace
	// to use for pulling the sipp image
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// Annotations added to the created jobs
	// +optional
	JobAnnotations map[string]string `json:"annotations,omitempty"`

	// ScenarioRef holds the fields to identify the scenario used for this run
	ScenarioRef *corev1.LocalObjectReference `json:"scenarioRef"`

	// CommandOverride allows to bypass all configuration fields
	// If set, all fields are ignored
	// +optional
	CommandOverride string `json:"commandOverride,omitempty"`
	// Destination
	// +optional
	Destination string `json:"destination,omitempty"`

	// Transport
	// See the -t parameter documentation
	// +optional
	Transport *Transport `json:"transport,omitempty"`

	// CallLength controls the length of calls
	// See the -d parameter documentation
	// +optional
	CallLength *int32 `json:"callLength,omitempty"`

	// ExitWhenCallsProcessed sets sipp to stop the test and exit
	// when 'calls' calls are processed
	// +optional
	ExitWhenCallsProcessed *bool `json:"exitWhenCallsProcessed,omitempty"`
}

// SippScenarioRunStatus defines the observed state of SippScenarioRun
type SippScenarioRunStatus struct {
	// The number of actively running sipp instance.
	// +optional
	Active int32 `json:"active,omitempty"`
	// The number of sipp instances which reached phase Succeeded.
	// +optional
	Succeeded int32 `json:"succeeded,omitempty"`
	// The number of sipp instances which reached phase Failed.
	// +optional
	Failed int32 `json:"failed,omitempty"`
}

// +kubebuilder:object:root=true

// SippScenarioRun is the Schema for the sippscenarioruns API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Active",type="integer",JSONPath=".status.active"
// +kubebuilder:printcolumn:name="Succeeded",type="integer",JSONPath=".status.succeeded"
// +kubebuilder:printcolumn:name="Failed",type="integer",JSONPath=".status.failed"
// +kubebuilder:resource:shortName={"ssr"}
type SippScenarioRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SippScenarioRunSpec   `json:"spec,omitempty"`
	Status SippScenarioRunStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SippScenarioRunList contains a list of SippScenarioRun
type SippScenarioRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SippScenarioRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SippScenarioRun{}, &SippScenarioRunList{})
}

// ChildResourceName returns the name of a child ressource using
// the scenario run's name
func (run SippScenarioRun) ChildResourceName(name string) string {
	return strings.TrimSuffix(strings.Join([]string{run.Name, name}, "-"), "-")
}

// ToSippArgs returns the Sipp Args from the Spec
// This function asserts that the Spec is clean (no unknown values)
func (run *SippScenarioRun) ToSippArgs() []string {
	if run.Spec.CommandOverride != "" {
		return strings.Split(run.Spec.CommandOverride, " ")
	}

	result := []string{}

	if run.Spec.ExitWhenCallsProcessed != nil && *run.Spec.ExitWhenCallsProcessed {
		result = append(result, "-m", "1")
	}

	if run.Spec.Transport != nil {
		result = append(result, run.TransportToSippArgs()...)
	}

	if run.Spec.CallLength != nil {
		result = append(result, "-d", strconv.FormatInt(int64(*run.Spec.CallLength), 10))
	}

	return result
}

// TransportToSippArgs returns Spec.Transport to Sipp args
// This function asserts that the transport is clean (no unknown values)
func (run *SippScenarioRun) TransportToSippArgs() []string {
	result := []string{"-t"}

	flag := ""

	if run.Spec.Transport.Compression != nil && *run.Spec.Transport.Compression {
		if run.Spec.Transport.Protocol == ProtocolUDP {
			flag += "c"
			switch run.Spec.Transport.Socket {
			case SocketOne:
				flag += "1"
			case SocketOnePerCall:
				flag += "n"
			}

			return append(result, flag)
		}
	}

	switch run.Spec.Transport.Protocol {
	case ProtocolTCP:
		flag += "t"
	case ProtocolUDP:
		flag += "u"
	case ProtocolTLS:
		flag += "l"
	}

	switch run.Spec.Transport.Socket {
	case SocketOne:
		flag += "1"
	case SocketOnePerCall:
		flag += "n"
	case SocketOnePerIP:
		flag += "i"
	}

	return append(result, flag)
}
