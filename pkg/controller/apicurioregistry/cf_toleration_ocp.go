package apicurioregistry

import (
	"reflect"

	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	ocp_apps "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ ControlFunction = &TolerationOcpCF{}

type TolerationOcpCF struct {
	ctx                         *Context
	deploymentConfigEntry       ResourceCacheEntry
	deploymentConfigEntryExists bool
	existingTolerations         []corev1.Toleration
	targetTolerations           []corev1.Toleration
}

func NewTolerationOcpCF(ctx *Context) ControlFunction {
	return &TolerationOcpCF{
		ctx:                         ctx,
		deploymentConfigEntry:       nil,
		deploymentConfigEntryExists: false,
		existingTolerations:         nil,
		targetTolerations:           nil,
	}
}

func (this *TolerationOcpCF) Describe() string {
	return "TolerationOcpCF"
}

func (this *TolerationOcpCF) Sense() {
	// Observation #1
	// Get the cached deploymentConfig
	this.deploymentConfigEntry, this.deploymentConfigEntryExists = this.ctx.GetResourceCache().Get(RC_KEY_DEPLOYMENT_OCP)

	if this.deploymentConfigEntryExists {
		// Observation #2
		// Get the existing tolerations
		this.existingTolerations = this.deploymentConfigEntry.GetValue().(*ocp_apps.DeploymentConfig).Spec.Template.Spec.Tolerations

		// Observation #3
		// Get the target tolerations
		if specEntry, exists := this.ctx.GetResourceCache().Get(RC_KEY_SPEC); exists {
			this.targetTolerations = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Tolerations
		}
	}
}

func (this *TolerationOcpCF) Compare() bool {
	// Condition #1
	// Deployment exists
	// Condition #2
	// Target toleration exists
	// Condition #3
	// Existing tolerations are different from target tolerations
	return this.deploymentConfigEntryExists &&
		!reflect.DeepEqual(this.existingTolerations, this.targetTolerations)
}

func (this *TolerationOcpCF) Respond() {
	// Response #1
	// Patch the resource
	this.deploymentConfigEntry.ApplyPatch(func(value interface{}) interface{} {
		deploymentConfig := value.(*ocp_apps.DeploymentConfig).DeepCopy()
		deploymentConfig.Spec.Template.Spec.Tolerations = this.targetTolerations
		return deploymentConfig
	})
}
