package resource

import (
	"github.com/pkg/errors"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/alexandrevilain/sipp-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	configPath = "/etc/jobconfig"
)

type JobBuilder struct {
	Instance *v1alpha1.SippScenarioRun
	Scenario *v1alpha1.SippScenario
	Scheme   *runtime.Scheme
}

func NewJobBuilder(builder *SippResourceBuilder) *JobBuilder {
	return &JobBuilder{
		Instance: builder.Instance,
		Scenario: builder.Scenario,
		Scheme:   builder.Scheme,
	}
}

func (b *JobBuilder) getLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      b.Instance.ChildResourceName("job"),
		"app.kubernetes.io/component": "job",
		"app.kubernetes.io/part-of":   "sipp-run",
	}
}

func (b *JobBuilder) getAnnotations() map[string]string {
	return b.Instance.Spec.JobAnnotations
}

func (b *JobBuilder) Build() (runtime.Object, error) {
	image := b.Instance.Spec.Image
	if image == "" {
		image = "ctaloi/sipp"
	}

	args := append([]string{}, b.Instance.ToSippArgs()...)
	args = append(args, b.Scenario.ToSippArgs(configPath)...)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.Instance.ChildResourceName("job"),
			Namespace:   b.Instance.Namespace,
			Labels:      b.getLabels(),
			Annotations: b.getAnnotations(),
		},
		Spec: batchv1.JobSpec{
			Parallelism: b.Instance.Spec.Parallelism,
			Completions: b.Instance.Spec.Parallelism,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:        b.Instance.ChildResourceName("job"),
					Namespace:   b.Instance.Namespace,
					Labels:      b.getLabels(),
					Annotations: b.getAnnotations(),
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: b.Instance.Spec.ImagePullSecrets,
					RestartPolicy:    "Never",
					Containers: []corev1.Container{
						{
							Name:  "sipp",
							Image: image,
							Args:  args,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "sipp-config",
									MountPath: configPath,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "sipp-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: b.Instance.ChildResourceName("configmap"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	err := controllerutil.SetControllerReference(b.Instance, job, b.Scheme)
	if err != nil {
		return job, errors.Wrap(err, "failed setting controller reference")
	}

	return job, nil
}

func (b *JobBuilder) Update(object runtime.Object) error {
	// Job should not be updated as its launched when created
	return nil
}
