package v1alpha1_test

import (
	"strings"
	"testing"

	"github.com/alexandrevilain/sipp-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestToSippArgs(t *testing.T) {
	scenario := &v1alpha1.SippScenario{
		Spec: v1alpha1.SippScenarioSpec{
			ScenarioFileContent: `<scenario name="Basic Sipstone UAC"></scenario>`,
			InjectValues: []string{
				`
				SEQUENTIAL
				#This line will be ignored
				Sarah;sipphone32
				`,
			},
		},
	}

	args := scenario.ToSippArgs("/etc/test")
	assert.Equal(t, "-sf /etc/test/scenario.xml -inf /etc/test/values_0.csv", strings.Join(args, " "))
}
