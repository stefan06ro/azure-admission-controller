package cluster

import (
	"context"
	"fmt"

	aeV3conditions "github.com/giantswarm/apiextensions/v3/pkg/conditions"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"
)

func (m *CreateMutator) getDefaultConditions(ctx context.Context) capiv1alpha3.Conditions {
	var conditions []capiv1alpha3.Condition

	// Add Creating condition
	creatingCondition := capiconditions.TrueCondition(aeV3conditions.CreatingCondition)
	conditions = append(conditions, *creatingCondition)
	m.logAddingDefaultCondition(ctx, creatingCondition)

	return conditions
}

func (m *CreateMutator) logAddingDefaultCondition(ctx context.Context, condition *capiv1alpha3.Condition) {
	m.logger.LogCtx(
		ctx,
		"level", "debug",
		"message", fmt.Sprintf("will set default condition %s=%s, Severity=%s, Reason=%s", condition.Type, condition.Status, condition.Severity, condition.Reason))
}
