package store

import (
	"context"
	"fmt"
	"strings"

	"testing"

	"github.com/enrayga/omc-o2ims/internal/operator/resource"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	fakeClientSet "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestNewK8sStore(t *testing.T) {

	t.Run("Test when crd is an empty string",
		func(t *testing.T) {
			ctx := context.Background()
			clientset := fakeClientSet.NewSimpleClientset()
			dynamicClient := fake.NewSimpleDynamicClient(runtime.NewScheme())
			store, err := NewK8sStore[*resource.MockResource](clientset,
				dynamicClient, ctx, "", "default")
			_ = store

			expectedErr := "CRDDefinition cannot be empty string"
			if err.Error() != expectedErr {
				t.Errorf("Expected error: %v, got: %v", expectedErr, err)
			}
		})

	t.Run("Test when namespace is an empty string",
		func(t *testing.T) {
			ctx := context.Background()
			clientset := fakeClientSet.NewSimpleClientset()
			dynamicClient := fake.NewSimpleDynamicClient(runtime.NewScheme())
			store, err := NewK8sStore[*resource.MockResource](clientset,
				dynamicClient, ctx, "DummyCRDDefinition", "")
			_ = store

			expectedErr := "namespace cannot be empty string"
			if err.Error() != expectedErr {
				t.Errorf("Expected error: %v, got: %v", expectedErr, err)
			}
		})

	t.Run("Test when clientset is nil",
		func(t *testing.T) {
			ctx := context.Background()
			//clientset := nil
			dynamicClient := fake.NewSimpleDynamicClient(runtime.NewScheme())
			store, err := NewK8sStore[*resource.MockResource](nil,
				dynamicClient, ctx, "DummyCRDDefinition", "default")
			_ = store

			expectedErr := "clientset cannot be nil"
			if err.Error() != expectedErr {
				t.Errorf("Expected error: %v, got: %v", expectedErr, err)
			}
		})

	t.Run("Test when dynamicclient is nil",
		func(t *testing.T) {
			ctx := context.Background()
			clientset := fakeClientSet.NewSimpleClientset()
			store, err := NewK8sStore[*resource.MockResource](clientset,
				nil, ctx, "dafds", "default")
			_ = store

			expectedErr := "dynamicClient cannot be nil"
			if err.Error() != expectedErr {
				t.Errorf("Expected error: %v, got: %v", expectedErr, err)
			}
		})

	t.Run("Test when ctx is nil",
		func(t *testing.T) {
			clientset := fakeClientSet.NewSimpleClientset()
			dynamicClient := fake.NewSimpleDynamicClient(runtime.NewScheme())
			store, err := NewK8sStore[*resource.MockResource](clientset,
				dynamicClient, nil, "dafds", "default")
			_ = store

			expectedErr := "context cannot be nil"
			if err.Error() != expectedErr {
				t.Errorf("Expected error: %v, got: %v", expectedErr, err)
			}
		})

	t.Run("Test when CRD is a valid json file",
		func(t *testing.T) {
			ctx := context.Background()
			clientset := fakeClientSet.NewSimpleClientset()
			crd := `{
  "apiVersion": "apiextensions.k8s.io/v1",
  "kind": "CustomResourceDefinition",
  "metadata": {
    "name": "myresources.example.com"
  },
  "spec": {
    "group": "example.com",
    "names": {
      "kind": "MyResource",
      "listKind": "MyResourceList",
      "plural": "myresources",
      "singular": "myresource"
    },
    "scope": "Namespaced",
    "versions": [
      {
        "name": "v1",
        "served": true,
        "storage": true,
        "schema": {
          "openAPIV3Schema": {
            "type": "object",
            "properties": {
              "spec": {
                "type": "object",
                "properties": {
                  "name": {
                    "type": "string"
                  },
                  "replicas": {
                    "type": "integer"
                  }
                }
              }
            }
          }
        }
      }
    ]
  }
}
`
			dynamicClient := fake.NewSimpleDynamicClient(runtime.NewScheme())

			store, err := NewK8sStore[*resource.MockResource](clientset,
				dynamicClient, ctx, string(crd), "default")
			_ = store
			if err != nil {
				t.Errorf("Error creating store: %v", err)
				return
			} else {
				fmt.Printf("Store created successfully\n")
			}

		})
	t.Run("Test when CRD is a invalid json file",
		func(t *testing.T) {
			ctx := context.Background()
			clientset := fakeClientSet.NewSimpleClientset()
			crd := `{
  "apiVersion": "apiextensions.k8s.io/v1",
  "kind": "CustomResourceDefinition",
  "metadata": DELETED {
    "name": "myresources.example.com"
  },
  "spec": {
    "group": "example.com",
    "names": {
      "kind": "MyResource",
      "listKind": "MyResourceList",
      "plural": "myresources",
      "singular": "myresource"
    },
    "scope": "Namespaced",
    "versions": [
      {
        "name": "v1",
        "served": true,
        "storage": true,
        "schema": {
          "openAPIV3Schema": {
            "type": "object",
            "properties": {
              "spec": {
                "type": "object",
                "properties": {
                  "name": {
                    "type": "string"
                  },
                  "replicas": {
                    "type": "integer"
                  }
                }
              }
            }
          }
        }
      }
    ]
  }
}
`
			dynamicClient := fake.NewSimpleDynamicClient(runtime.NewScheme())

			store, err := NewK8sStore[*resource.MockResource](clientset,
				dynamicClient, ctx, string(crd), "default")
			_ = store

			expected := "line 3: did not find expected ',' or '}'"
			if err != nil && !strings.Contains(err.Error(), expected) {
				t.Errorf("expected %v but gotUnexpected error: %v", expected, err)
				return
			}

		})

	t.Run("Create CRD from valid yaml", func(t *testing.T) {

		crd := `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: dummyprovisionings.o2ims.example.com
spec:
  group: o2ims.example.com
  names:
    kind: DummyProvisioning
    listKind: DummyProvisioningList
    plural: dummyprovisionings
    singular: dummyprovisioning
    shortNames:
      - dp
  scope: Namespaced
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                name:
                  type: string
                replicas:
                  type: integer
`
		ctx := context.Background()
		clientset := fakeClientSet.NewSimpleClientset()
		dynamicClient := fake.NewSimpleDynamicClient(runtime.NewScheme())

		store, err := NewK8sStore[*resource.MockResource](clientset,
			dynamicClient, ctx, crd, "default")

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if store.Crd.Name != "dummyprovisionings.o2ims.example.com" {
			t.Errorf("Expected name to be dummyprovisionings.o2ims.example.com but got %s", store.Crd.Name)
		}
	})

	t.Run("Create CRD from yaml with error", func(t *testing.T) {

		crd := `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata REMOVED COLON
  name: dummyprovisionings.o2ims.example.com 
spec:
  group: o2ims.example.com
  names:
    kind: DummyProvisioning
    listKind: DummyProvisioningList
    plural: dummyprovisionings
    singular: dummyprovisioning
    shortNames:
      - dp
  scope: Namespaced
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                name:
                  type: string
                replicas:
                  type: integer
`
		ctx := context.Background()
		clientset := fakeClientSet.NewSimpleClientset()
		dynamicClient := fake.NewSimpleDynamicClient(runtime.NewScheme())

		store, err := NewK8sStore[*resource.MockResource](clientset,
			dynamicClient, ctx, crd, "default")

		_ = store

		//fmt.Printf("Error: %v\n", err)
		expected := "line 5: could not find expected ':'"
		if err != nil && !strings.Contains(err.Error(), expected) {
			t.Errorf("expected %v but got Unexpected error: %v", expected, err)
		}
	})

	t.Run("Test when get CRD failed for NotFound reason (Leads to CRD creation)", func(t *testing.T) {
		crd := `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: dummyprovisionings.o2ims.example.com
spec:
  group: o2ims.example.com
  names:
    kind: DummyProvisioning
    listKind: DummyProvisioningList
    plural: dummyprovisionings
    singular: dummyprovisioning
    shortNames:
      - dp
  scope: Namespaced
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                name:
                  type: string
                replicas:
                  type: integer
`
		ctx := context.Background()
		clientset := fakeClientSet.NewSimpleClientset()

		res := schema.GroupResource{Group: "o2ims.example.com", Resource: "myresources"}
		name := "dummyprovisionings.o2ims.example.com"

		//Note : this is the default behviour of the fake clientset to return notfound
		clientset.Fake.PrependReactor("get", "customresourcedefinitions", func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, apierrors.NewNotFound(res, name)
		})
		dynamicClient := fake.NewSimpleDynamicClient(runtime.NewScheme())

		store, err := NewK8sStore[*resource.MockResource](clientset,
			dynamicClient, ctx, crd, "default")

		_ = store

		if err != nil {
			t.Errorf("Expected no error but got error Unexpected error: %v", err)
		}

	})
	t.Run("Test when get CRD fails for unkown reason (not NotFound) and leads to error out", func(t *testing.T) {
		crd := `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: dummyprovisionings.o2ims.example.com
spec:
  group: o2ims.example.com
  names:
    kind: DummyProvisioning
    listKind: DummyProvisioningList
    plural: dummyprovisionings
    singular: dummyprovisioning
    shortNames:
      - dp
  scope: Namespaced
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                name:
                  type: string
                replicas:
                  type: integer
`
		ctx := context.Background()
		clientset := fakeClientSet.NewSimpleClientset()
		expected := "Get  simulated  failure"

		error_var := fmt.Errorf("%s", expected)
		clientset.Fake.PrependReactor("get", "customresourcedefinitions", func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, error_var
		})
		dynamicClient := fake.NewSimpleDynamicClient(runtime.NewScheme())

		store, err := NewK8sStore[*resource.MockResource](clientset,
			dynamicClient, ctx, crd, "default")

		_ = store

		if !(err != nil && strings.Contains(err.Error(), expected)) {
			t.Errorf("expected %v but got Unexpected error: %v", expected, err)
			return
		}

	})

	t.Run("Test when CRD creation failed", func(t *testing.T) {

		crd := `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: dummyprovisionings.o2ims.example.com
spec:
  group: o2ims.example.com
  names:
    kind: DummyProvisioning
    listKind: DummyProvisioningList
    plural: dummyprovisionings
    singular: dummyprovisioning
    shortNames:
      - dp
  scope: Namespaced
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                name:
                  type: string
                replicas:
                  type: integer
`
		ctx := context.Background()
		clientset := fakeClientSet.NewSimpleClientset()
		expected := "Creation failed simulated creation failure"

		error_var := fmt.Errorf("%s", expected)
		// Add a reactor to simulate failure on CRD creation
		clientset.Fake.PrependReactor("create", "customresourcedefinitions", func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, error_var
		})
		dynamicClient := fake.NewSimpleDynamicClient(runtime.NewScheme())

		store, err := NewK8sStore[*resource.MockResource](clientset,
			dynamicClient, ctx, crd, "default")

		_ = store

		if !(err != nil && strings.Contains(err.Error(), expected)) {
			t.Errorf("expected %v but got Unexpected error: %v", expected, err)
		}

	})

	t.Run("Create CRD when it is already present", func(t *testing.T) {

		crd := `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: mockresources.o2ims.provisioning.oran.org
spec:
  group: o2ims.provisioning.oran.org
  names:
    kind: MockResource
    plural: mockresources
    shortNames:
      - mr
    singular: mockresource
  scope: Namespaced
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                name:
                  type: string
                replicas:
                  type: integer
`
		ctx := context.Background()
		clientset := fakeClientSet.NewSimpleClientset()

		expected_crd := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "myresources.example.com",
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: "example.com",
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Kind:     "MyResource",
					ListKind: "MyResourceList",
					Plural:   "myresources",
					Singular: "myresource",
				},
				Scope: apiextensionsv1.NamespaceScoped,
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name:    "v1",
						Served:  true,
						Storage: true,
					},
				},
			},
		}

		clientset.Fake.PrependReactor("get", "customresourcedefinitions", func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, expected_crd, nil
		})

		dynamicClient := fake.NewSimpleDynamicClient(runtime.NewScheme())

		store, err := NewK8sStore[*resource.MockResource](clientset,
			dynamicClient, ctx, crd, "default")

		_ = store

		if err != nil {
			t.Errorf("Expected no error but got error Unexpected error: %v", err)
		}

	})
}

func TestList(t *testing.T) {
	clientset := fakeClientSet.NewSimpleClientset()

	ctx := context.Background()

	crdDefinition := `
{
    "apiVersion": "apiextensions.k8s.io/v1",
    "kind": "CustomResourceDefinition",
    "metadata": {
        "name": "myresources.example.com"
    },
    "spec": {
        "group": "example.com",
        "names": {
            "kind": "MyResource",
            "listKind": "MyResourceList",
            "plural": "myresources",
            "singular": "myresource"
        },
        "scope": "Namespaced",
        "versions": [
            {
                "name": "v1",
                "served": true,
                "storage": true
            }
        ]
    }
}
`

	expected_crd := &apiextensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: "apiextensions.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "myresources.example.com",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "example.com",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:     "MyResource",
				ListKind: "MyResourceList",
				Plural:   "myresources",
				Singular: "myresource",
			},
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
		},
	}

	clientset.Fake.PrependReactor("get", "customresourcedefinitions", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, expected_crd, nil
	})

	dynamicClient := fake.NewSimpleDynamicClient(runtime.NewScheme())

	store, err := NewK8sStore[*resource.MockResource](clientset,
		dynamicClient, ctx, crdDefinition, "default")

	_ = store

	if err != nil {
		t.Errorf("Expected no error but got error Unexpected error: %v", err)
	}

	resources, err := store.List()

	if err != nil {
		t.Errorf("Expected no error but got error Unexpected error: %v", err)
	}

	if len(resources) != 0 {
		t.Errorf("Expected no resources but got %d resources", len(resources))
	}
}

func TestReconcileList(t *testing.T) {
	ctx := context.Background()
	clientset := fakeClientSet.NewSimpleClientset()

	crdDefinition := `
	{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind": "CustomResourceDefinition",
		"metadata": {
			"name": "myresources.example.com"
		},
		"spec": {
			"group": "example.com",
			"names": {
				"kind": "MyResource",
				"listKind": "MyResourceList",
				"plural": "myresources",
				"singular": "myresource"
			},
			"scope": "Namespaced",
			"versions": [
				{
					"name": "v1",
					"served": true,
					"storage": true
				}
			]
		}
	}
	`
	scheme := runtime.NewScheme()
	expected_crd := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myresources.example.com",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "example.com",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:     "MyResource",
				ListKind: "MyResourceList",
				Plural:   "myresources",
				Singular: "myresource",
			},
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
		},
	}

	// Create list GVK for registration
	gvk := schema.GroupVersionKind{
		Group:   "example.com",
		Version: "v1",
		Kind:    "MyResourceList",
	}

	gvr := schema.GroupVersionResource{
		Group:    "example.com",
		Version:  "v1",
		Resource: "myresources",
	}

	listKinds := map[schema.GroupVersionResource]string{
		gvr: gvk.Kind,
	}
	dynamicClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds)

	clientset.Fake.PrependReactor("get", "customresourcedefinitions",
		func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, expected_crd, nil
		})

	store, _ := NewK8sStore[*resource.MockResource](clientset,
		dynamicClient, ctx, crdDefinition, "default")

	mockResource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "example.com/v1",
			"kind":       "MyResource",
			"metadata": map[string]interface{}{
				"name":      "test-resource",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"size":    "3",
				"message": "test message",
			},
		},
	}

	list_out := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "example.com/v1",
			"kind":       "MyResourceList",
			"metadata": map[string]interface{}{
				"resourceVersion": "1",
			},
		},
		Items: []unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"apiVersion": "example.com/v1",
					"kind":       "MyResource",
					"metadata": map[string]interface{}{
						"name":      "test-resource",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"size":    "3",
						"message": "test message",
					},
				},
			},
			{
				Object: map[string]interface{}{
					"apiVersion": "example.com/v1",
					"kind":       "MyResource",
					"metadata": map[string]interface{}{
						"name":      "other-resource",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"size":    "4",
						"message": "other message",
					},
				},
			},
		},
	}
	_ = mockResource
	dynamicClient.PrependReactor("list", "myresources",
		func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			//list := list_out
			//fmt.Printf("list: %v\n", list)
			return true, list_out, nil
		},
	)

	dynamicClient.PrependReactor("get", "myresources",
		func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			getAction := action.(k8stesting.GetAction)

			for i, item := range list_out.Items {
				resource_out := item
				// Match the resource name
				if getAction.GetName() == resource_out.GetName() {
					return true, &list_out.Items[i], nil
				}
			}
			return true, nil, fmt.Errorf("resource %s not found", getAction.GetName())
		},
	)

	dynamicClient.PrependReactor("update", "myresources",
		func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			updateAction, ok := action.(k8stesting.UpdateAction)
			if !ok {
				return false, nil, fmt.Errorf("unexpected action type")
			}

			obj := updateAction.GetObject().(*unstructured.Unstructured)
			if err != nil {
				return false, nil, fmt.Errorf("failed to access object meta: %v", err)
			}

			// Simulate setting the finalizer
			finalizers := obj.GetFinalizers()
			//finalizers = append(finalizers, "example.com/finalizer")
			obj.SetFinalizers(finalizers)

			// Return the modified object as the result of the "update" operation
			return true, obj, nil
		})

	store.ReconcileList()

	//check if two resources are in the store
	if len(store.currentResources) != 2 {
		t.Errorf("expected 2 resources in the store, got %d", len(store.currentResources))
	}

	//Check if there is no change
	store.ReconcileList()
	if len(store.currentResources) != 2 {
		t.Errorf("expected 2 resources in the store, got %d", len(store.currentResources))
	}
}
