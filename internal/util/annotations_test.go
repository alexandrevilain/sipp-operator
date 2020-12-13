package util_test

import (
	"github.com/alexandrevilain/sipp-operator/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMergeAnnotations(t *testing.T) {
	annotations1 := map[string]string{
		"app.kubernetes.io/app":       "my_app",
		"app.kubernetes.io/component": "my_component",
	}

	annotations2 := map[string]string{
		"app.kubernetes.io/app": "my_app_overrided",
		"my_custom_annotation":  "hey",
	}

	result := util.MergeAnnotations(annotations1, annotations2)

	assert.Equal(t, result["app.kubernetes.io/app"], "my_app_overrided")
	assert.Equal(t, result["app.kubernetes.io/component"], "my_component")
	assert.Equal(t, result["my_custom_annotation"], "hey")
}
