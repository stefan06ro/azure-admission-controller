package filter

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	providerv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint:staticcheck
	"sigs.k8s.io/yaml"
)

const (
	// legacyRelease is a Giant Swarm release which contains azure-operator and
	// it is expected that azure-operator will reconcile CRs that belong to
	// clusters with this release. You can find Release CR manifest in testdata
	// directory.
	legacyRelease = "14.1.4"

	// capiRelease is a new Giant Swarm CAPI-style release which does not
	// contain azure-operator and it is expected that azure-operator will not
	// reconcile CRs that belong to clusters with this release. You can find
	// Release CR manifest in testdata directory.
	capiRelease = "20.0.0-v1alpha3"

	defaultTestCRName      = "hello"
	defaultTestCRNamespace = "org-test"
)

func Test_IsObjectReconciledByLegacyRelease(t *testing.T) {
	testCases := []struct {
		name           string
		inputCR        object
		ownerCluster   object
		expectedResult bool
	}{
		//
		// Test cases where the CR from "legacy" Giant Swarm release is processed
		//

		// Cluster
		{
			name:           "Cluster with release label is processed",
			inputCR:        cluster().withReleaseVersionLabel(legacyRelease),
			expectedResult: true,
		},
		// AzureCluster
		{
			name:           "AzureCluster with release label is processed",
			inputCR:        azureCluster().withReleaseVersionLabel(legacyRelease),
			expectedResult: true,
		},
		{
			name:           "AzureCluster with cluster name label is processed",
			inputCR:        azureCluster().withClusterNameLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(legacyRelease).object,
			expectedResult: true,
		},
		{
			name:           "AzureCluster with cluster ID label is processed",
			inputCR:        azureCluster().withClusterIDLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(legacyRelease).object,
			expectedResult: true,
		},
		// MachinePool
		{
			name:           "MachinePool with release label is processed",
			inputCR:        machinePool().withReleaseVersionLabel(legacyRelease),
			expectedResult: true,
		},
		{
			name:           "MachinePool with cluster name label is processed",
			inputCR:        machinePool().withClusterNameLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(legacyRelease).object,
			expectedResult: true,
		},
		{
			name:           "MachinePool with cluster ID label is processed",
			inputCR:        machinePool().withClusterIDLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(legacyRelease).object,
			expectedResult: true,
		},
		// AzureMachinePool
		{
			name:           "AzureMachinePool with release label is processed",
			inputCR:        azureMachinePool().withReleaseVersionLabel(legacyRelease),
			expectedResult: true,
		},
		{
			name:           "AzureMachinePool with cluster name label is processed",
			inputCR:        azureMachinePool().withClusterNameLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(legacyRelease).object,
			expectedResult: true,
		},
		{
			name:           "AzureMachinePool with cluster ID label is processed",
			inputCR:        azureMachinePool().withClusterIDLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(legacyRelease).object,
			expectedResult: true,
		},
		// AzureMachine
		{
			name:           "AzureMachine with release label is processed",
			inputCR:        azureMachine().withReleaseVersionLabel(legacyRelease),
			expectedResult: true,
		},
		{
			name:           "AzureMachine with cluster name label is processed",
			inputCR:        azureMachine().withClusterNameLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(legacyRelease).object,
			expectedResult: true,
		},
		{
			name:           "AzureMachine with cluster ID label is processed",
			inputCR:        azureMachine().withClusterIDLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(legacyRelease).object,
			expectedResult: true,
		},

		//
		// Test cases where the CR from CAPI release is not processed
		//

		// Cluster
		{
			name:           "Cluster without release label is not processed",
			inputCR:        cluster(),
			expectedResult: false,
		},
		{
			name:           "Cluster with release label is not processed",
			inputCR:        cluster().withReleaseVersionLabel(capiRelease),
			expectedResult: false,
		},
		// AzureCluster
		{
			name:           "AzureCluster without release label and without cluster name/id label is not processed",
			inputCR:        azureCluster(),
			expectedResult: false,
		},
		{
			name:           "AzureCluster with release label is not processed",
			inputCR:        azureCluster().withReleaseVersionLabel(capiRelease),
			expectedResult: false,
		},
		{
			name:           "AzureCluster with cluster name label is not processed",
			inputCR:        azureCluster().withClusterNameLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(capiRelease).object,
			expectedResult: false,
		},
		{
			name:           "AzureCluster with cluster ID label is not processed",
			inputCR:        azureCluster().withClusterIDLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(capiRelease).object,
			expectedResult: false,
		},
		// MachinePool
		{
			name:           "MachinePool without release label and without cluster name/id label is not processed",
			inputCR:        machinePool(),
			expectedResult: false,
		},
		{
			name:           "MachinePool with release label is not processed",
			inputCR:        machinePool().withReleaseVersionLabel(capiRelease),
			expectedResult: false,
		},
		{
			name:           "MachinePool with cluster name label is not processed",
			inputCR:        machinePool().withClusterNameLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(capiRelease).object,
			expectedResult: false,
		},
		{
			name:           "MachinePool with cluster ID label is not processed",
			inputCR:        machinePool().withClusterIDLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(capiRelease).object,
			expectedResult: false,
		},
		// AzureMachinePool
		{
			name:           "AzureMachinePool without release label and without cluster name/id label is not processed",
			inputCR:        azureMachinePool(),
			expectedResult: false,
		},
		{
			name:           "AzureMachinePool with release label is not processed",
			inputCR:        azureMachinePool().withReleaseVersionLabel(capiRelease),
			expectedResult: false,
		},
		{
			name:           "AzureMachinePool with cluster name label is not processed",
			inputCR:        azureMachinePool().withClusterNameLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(capiRelease).object,
			expectedResult: false,
		},
		{
			name:           "AzureMachinePool with cluster ID label is not processed",
			inputCR:        azureMachinePool().withClusterIDLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(capiRelease).object,
			expectedResult: false,
		},
		// AzureMachine
		{
			name:           "AzureMachine without release label and without cluster name/id label is not processed",
			inputCR:        azureMachine(),
			expectedResult: false,
		},
		{
			name:           "AzureMachine with release label is not processed",
			inputCR:        azureMachine().withReleaseVersionLabel(capiRelease),
			expectedResult: false,
		},
		{
			name:           "AzureMachine with cluster name label is not processed",
			inputCR:        azureMachine().withClusterNameLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(capiRelease).object,
			expectedResult: false,
		},
		{
			name:           "AzureMachine with cluster ID label is not processed",
			inputCR:        azureMachine().withClusterIDLabel(),
			ownerCluster:   cluster().withReleaseVersionLabel(capiRelease).object,
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.name)
			ctx := context.Background()
			ctrlClient := newFakeClient()
			loadReleases(t, ctrlClient)

			clusterGetter := func(metav1.Object) capi.Cluster {
				if tc.ownerCluster == nil {
					return capi.Cluster{}
				}

				cluster, ok := tc.ownerCluster.(*capi.Cluster)
				if !ok {
					t.Fatalf("Owner cluster is not a Cluster CR, check test inputs, got %T, expected *Cluster.", tc.ownerCluster)
				}

				return *cluster
			}

			result, err := IsObjectReconciledByLegacyRelease(ctx, ctrlClient, tc.inputCR, clusterGetter)
			if err != nil {
				t.Fatalf("Error when calling IsObjectReconciledByLegacyRelease: %#v", err)
			}

			if result != tc.expectedResult {
				t.Fatal()
			}
		})
	}
}

func loadReleases(t *testing.T, client client.Client) {
	testReleasesToLoad := []string{legacyRelease, capiRelease}

	for _, releaseVersion := range testReleasesToLoad {
		var err error
		fileName := fmt.Sprintf("release-v%s.yaml", releaseVersion)

		var bs []byte
		{
			bs, err = ioutil.ReadFile(filepath.Join("testdata", fileName))
			if err != nil {
				t.Fatalf("failed to create object from input file %s: %#v", fileName, err)
			}
		}

		// First parse kind.
		typeMeta := &metav1.TypeMeta{}
		err = yaml.Unmarshal(bs, typeMeta)
		if err != nil {
			t.Fatalf("failed to create object from input file %s: %#v", fileName, err)
		}

		if typeMeta.Kind != "Release" {
			t.Fatalf("failed to create object from input file %s, the file does not contain Release CR", fileName)
		}

		var release releasev1alpha1.Release
		err = yaml.Unmarshal(bs, &release)
		if err != nil {
			t.Fatalf("failed to create object from input file %s: %#v", fileName, err)
		}

		err = client.Create(context.Background(), &release)
		if err != nil {
			t.Fatalf("failed to create object from input file %s: %#v", fileName, err)
		}
	}
}

func newFakeClient() client.Client {
	scheme := runtime.NewScheme()

	err := capi.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	err = capz.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	err = corev1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	err = providerv1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	err = releasev1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	return fake.NewFakeClientWithScheme(scheme)
}

type object interface {
	metav1.Object
	runtime.Object
}

type objectWrapper struct {
	object
}

func (b *objectWrapper) withLabel(label, labelValue string) *objectWrapper {
	labels := b.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	labels[label] = labelValue
	b.SetLabels(labels)
	return b
}

func (b *objectWrapper) withClusterNameLabel() *objectWrapper {
	return b.withLabel(capi.ClusterLabelName, b.object.GetName())
}

func (b *objectWrapper) withClusterIDLabel() *objectWrapper {
	return b.withLabel(label.Cluster, b.object.GetName())
}

func (b *objectWrapper) withReleaseVersionLabel(releaseVersion string) *objectWrapper {
	return b.withLabel(label.ReleaseVersion, releaseVersion)
}

func cluster() *objectWrapper {
	return &objectWrapper{
		&capi.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: defaultTestCRNamespace,
				Name:      defaultTestCRName,
			},
		},
	}
}

func machinePool() *objectWrapper {
	return &objectWrapper{
		&capiexp.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: defaultTestCRNamespace,
				Name:      defaultTestCRName,
			},
		},
	}
}

func azureCluster() *objectWrapper {
	return &objectWrapper{
		&capz.AzureCluster{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: defaultTestCRNamespace,
				Name:      defaultTestCRName,
			},
		},
	}
}

func azureMachine() *objectWrapper {
	return &objectWrapper{
		&capz.AzureMachine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: defaultTestCRNamespace,
				Name:      defaultTestCRName,
			},
		},
	}
}

func azureMachinePool() *objectWrapper {
	return &objectWrapper{
		&capzexp.AzureMachinePool{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: defaultTestCRNamespace,
				Name:      defaultTestCRName,
			},
		},
	}
}
