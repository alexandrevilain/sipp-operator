package v1alpha1_test

import (
	"strings"
	"testing"

	"github.com/alexandrevilain/sipp-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/pointer"
)

func TestComputeArgsOverride(t *testing.T) {
	override := "-sf scenario.xml DEST_IP -s DEST_NUMBER"

	run := &v1alpha1.SippScenarioRun{
		Spec: v1alpha1.SippScenarioRunSpec{
			Destination:     "121.0.0.1",
			CommandOverride: override,
		},
	}

	// Transport and destination should be ignored, as command Override is set
	res := run.ToSippArgs()
	assert.Equal(t, override, strings.Join(res, " "))
}

func TestTransportToSippArgs(t *testing.T) {
	tests := []struct {
		Run      *v1alpha1.SippScenarioRun
		Expected []string
	}{
		{
			Run: &v1alpha1.SippScenarioRun{
				Spec: v1alpha1.SippScenarioRunSpec{
					Transport: &v1alpha1.Transport{
						Protocol:    "TCP",
						Socket:      "One",
						Compression: pointer.BoolPtr(false),
					},
				},
			},
			Expected: []string{"-t", "t1"},
		},
		{
			Run: &v1alpha1.SippScenarioRun{
				Spec: v1alpha1.SippScenarioRunSpec{
					Transport: &v1alpha1.Transport{
						Protocol:    "UDP",
						Socket:      "One",
						Compression: pointer.BoolPtr(true),
					},
				},
			},
			Expected: []string{"-t", "c1"},
		},
		{
			Run: &v1alpha1.SippScenarioRun{
				Spec: v1alpha1.SippScenarioRunSpec{
					Transport: &v1alpha1.Transport{
						Protocol:    "TLS",
						Socket:      "OnePerCall",
						Compression: pointer.BoolPtr(true),
					},
				},
			},
			Expected: []string{"-t", "ln"},
		},
	}

	for _, test := range tests {
		result := test.Run.TransportToSippArgs()
		assert.Equal(t, test.Expected, result)
	}
}
