package generic

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCAPIReleaseLabel(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentRelease string
		expectedResult bool
	}{
		{
			// Release label is not set
			name: "case 0",
			ctx:  context.Background(),

			currentRelease: "",
			expectedResult: false,
		},
		{
			// CAPI Release label is set
			name: "case 1",
			ctx:  context.Background(),

			currentRelease: "20.0.0-v1alpha3",
			expectedResult: true,
		},
		{
			// CAPI Release label is set
			name: "case 2",
			ctx:  context.Background(),

			currentRelease: "20.0.0",
			expectedResult: true,
		},
		{
			// GS Release label is set
			name: "case 3",
			ctx:  context.Background(),

			currentRelease: "14.0.0",
			expectedResult: false,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			object := newObjectWithRelease(nil, &tc.currentRelease)
			capi, err := IsCAPIRelease(object)
			if err != nil {
				t.Fatal(err)
			}
			// check if the result label is as expected
			if tc.expectedResult != capi {
				t.Fatalf("expected %v to be equal to %v", tc.expectedResult, capi)
			}
		})
	}
}

func newObjectWithRelease(clusterID *string, release *string) metav1.Object {
	obj := &GenericObject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ab123",
			Namespace: "default",
			Labels: map[string]string{
				"azure-operator.giantswarm.io/version": "5.0.0",
				"cluster.x-k8s.io/cluster-name":        "ab123",
				"cluster.x-k8s.io/control-plane":       "true",
				"giantswarm.io/machine-pool":           "ab123",
			},
		},
	}

	if clusterID != nil {
		obj.Labels[label.Cluster] = *clusterID
	}

	if release != nil {
		obj.Labels[label.ReleaseVersion] = *release
	}

	return obj
}
