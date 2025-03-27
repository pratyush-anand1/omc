package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/enrayga/omc-o2ims/internal/operator/resource"
	"gopkg.in/yaml.v2"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const finalizerName = "o2ims.provisioning.oran.org.omc.v1alpha1"

var (
	InfoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarnLogger  = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

type ClientsetInterface interface {
	ApiextensionsV1() apiextensionsv1client.ApiextensionsV1Interface
}

type K8sStore[T resource.Resource] struct {
	// Crd represents the CustomResourceDefinition.
	//
	// It is first passed by the user and once the CRD is
	// created, it will be stored in this variable.
	//
	// Also, isCrdGenerated is set to true when the CRD is generated.
	Crd *apiextensionsv1.CustomResourceDefinition
	//isCrdGenerated bool

	// nameSpace represents the namespace in which the CRD is defined.
	nameSpace string

	// ApiClientSet represents the clientset for managing CRD definitions.
	clientset ClientsetInterface

	// dynamicClient represents the dynamic client for working with unknown
	// resource types.
	dynamicClient dynamic.Interface

	// list of monitored and deleting resources
	currentResources []T // List of resources that are currently present

}

type KubernetesInfo struct {
	// KubeConfig path to kubeconfig file
	KubeConfig string
	// Context is the kubeconfig context to use
	Context string
	// InCluster determines whether to use in-cluster config
	InCluster bool
}

func isJSON(str string) bool {
	var js json.RawMessage
	err := json.Unmarshal([]byte(str), &js)
	return (err == nil)
}

// FIXME
// Unmarshaling YAML did not work correctly so had to convert it to JSON
// now yet sure what the below code does but works so copy pasted for now :)
// Copied from Stack overflow
// https://stackoverflow.com/questions/40737122/convert-yaml-to-json-without-struct
func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}

func createCRDFromYaml(yamlData []byte) (apiextensionsv1.CustomResourceDefinition,
	error) {
	var yamlMap interface{}
	Crd := apiextensionsv1.CustomResourceDefinition{}
	err := yaml.Unmarshal(yamlData, &yamlMap)
	if err != nil {
		return Crd, err
	}
	//Hack :)
	yamlMap = convert(yamlMap)
	jsonData, err := json.Marshal(yamlMap)
	if err != nil {
		return Crd, err
	}
	err = json.Unmarshal(jsonData, &Crd)
	if err != nil {
		return Crd, err
	}
	return Crd, nil
}

func createCRDFromJSON(jsonData []byte) (apiextensionsv1.CustomResourceDefinition, error) {
	Crd := apiextensionsv1.CustomResourceDefinition{}
	err := json.Unmarshal(jsonData, &Crd)
	if err != nil {
		return Crd, err
	}
	return Crd, nil
}

func createCRD(data []byte) (*apiextensionsv1.CustomResourceDefinition, error) {
	if isJSON(string(data)) {
		crd, err := createCRDFromJSON(data)
		if err != nil {
			return nil, err
		}
		return &crd, nil
	} else {
		crd, err := createCRDFromYaml(data)
		if err != nil {
			return nil, err
		}
		return &crd, nil
	}
}

func createCRDIfNotExist(clientSet ClientsetInterface,
	namespace string,
	crdDefinition string) (*apiextensionsv1.CustomResourceDefinition, error) {
	// FIXME
	_ = namespace
	crd, err := createCRD([]byte(crdDefinition))
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	crd2, err := clientSet.ApiextensionsV1().CustomResourceDefinitions().Get(ctx,
		crd.Name, metav1.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			crd2, err = clientSet.ApiextensionsV1().CustomResourceDefinitions().Create(ctx,
				crd, metav1.CreateOptions{})
			if err == nil {
				InfoLogger.Printf("CRD %s Creation Succeeded\n", crd.Name)
				return crd2, nil
			} else {
				ErrorLogger.Printf("CRD %s Creation failed %v\n", crd.Name, err)
				return crd2, err
			}
		} else {
			ErrorLogger.Printf("CRD %s Unknown Error %v\n", crd.Name, err)
			return nil, err
		}
	}
	InfoLogger.Printf("CRD %s  Already Present !\n", crd.Name)
	return crd2, err
}

func NewK8sStore[T resource.Resource](
	clientset ClientsetInterface,
	dynamicClient dynamic.Interface,
	ctx context.Context,
	CRDDefinition string,
	namespace string) (*K8sStore[T], error) {

	k8sStore := K8sStore[T]{}
	k8sStore.clientset = clientset

	k8sStore.dynamicClient = dynamicClient
	k8sStore.nameSpace = namespace

	if clientset == nil {
		return nil, fmt.Errorf("clientset cannot be nil")
	}

	if dynamicClient == nil {
		return nil, fmt.Errorf("dynamicClient cannot be nil")
	}

	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if CRDDefinition == "" {
		return nil, fmt.Errorf("CRDDefinition cannot be empty string")
	}

	if namespace == "" {
		return nil, fmt.Errorf("namespace cannot be empty string")
	}

	var err error

	if k8sStore.Crd, err = createCRDIfNotExist(clientset,
		namespace,
		CRDDefinition); err != nil {
		return nil, err
	}
	return &k8sStore, err
}

func (s *K8sStore[T]) List() ([]T, error) {
	return s.currentResources, nil
}

// ReconcileList implements Store.ReconcileList
func (s *K8sStore[T]) ReconcileList() error {
	const (
		New int = iota
		Current
		Deleting
	)

	gvr := schema.GroupVersionResource{
		Group:    s.Crd.Spec.Group,
		Version:  s.Crd.Spec.Versions[0].Name,
		Resource: s.Crd.Spec.Names.Plural,
	}

	//FIXME
	//If cluster-scoped, do not use .Namespace().
	//If namespaced
	//list, err := s.dynamicClient.Resource(gvr).Namespace("default").List(context.TODO(), metav1.ListOptions{})

	unstructuredList, err := s.dynamicClient.
		Resource(gvr).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		// FIX ME we come here at two situation
		// 1.  CRD is not ready yet (Okay)
		// 2. CRD is deleted (Oops) we are out of sync as someone r
		// deleted CRDis too and we are out of sync and we will not
		// be able to reconcile ever .
		ErrorLogger.Printf("List error: %v\n", err)
		return err

	}

	for _, item := range unstructuredList.Items {
		unstructuredItem := item
		name, found, err := unstructured.NestedString(unstructuredItem.Object, "metadata", "name")
		//fmt.Printf("List:%T  %v, %v, %v\n", unstructuredItem, name, found, err)
		if !found || err != nil {
			break
		}

		//name and id are same
		state := New
		var existingRes T
		for _, res := range s.currentResources {
			if res.GetID() == name {
				existingRes = res
				state = Current

				// also check if it is being deleted
				deletionTimestamp, found, err := unstructured.NestedString(unstructuredItem.Object, "metadata", "deletionTimestamp")
				if found && deletionTimestamp != "" && err == nil {
					state = Deleting
				}
				break
			}
		}
		switch state {
		case New:
			InfoLogger.Println("New Resource: ", name)
			var zero T
			newResource := zero.GetNew()
			typedResource := newResource.(T)
			objectMap := unstructuredItem.Object
			_ = typedResource.SetInitFields(name, objectMap)

			err = s.ModifyFinalizer(name, true)
			if err != nil {
				ErrorLogger.Printf("add Finalizer error: %v\n", err)
			}
			//All is well update in the new to be added resoruce
			s.currentResources = append(s.currentResources, typedResource)
		case Current:
			InfoLogger.Println("Current Resource: ", name)
			objectMap := unstructuredItem.Object
			changed, err := existingRes.Compare(name, objectMap, true)
			if err != nil || changed {
				InfoLogger.Printf("changed: %v, err: %v\n", changed, err)
			}
		case Deleting:
			InfoLogger.Println("Deleteing Resource: ", name)
			deleting := existingRes.GetDeleteFlag()
			if !deleting {
				err = existingRes.SetDeleteFlag()
				if err != nil {
					ErrorLogger.Printf("Initiating delete: %v\n", err)
				}
			}
			continue
		}
	}
	return nil
}

func containsFinalizer(finalizers []string, finalizerName string) bool {
	for _, f := range finalizers {
		if f == finalizerName {
			return true
		}
	}
	return false
}

func removeFinalizer(finalizers []string, finalizerName string) []string {
	result := make([]string, 0)
	for _, f := range finalizers {
		if f != finalizerName {
			result = append(result, f)
		}
	}
	return result
}

func (s *K8sStore[T]) ModifyFinalizer(name string, add bool) error {
	gvr := schema.GroupVersionResource{
		Group:    s.Crd.Spec.Group,
		Version:  s.Crd.Spec.Versions[0].Name,
		Resource: s.Crd.Spec.Names.Plural,
	}

	crdInstance, err := s.dynamicClient.Resource(gvr).
		Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get resource: %v", err)
	}

	finalizers := crdInstance.GetFinalizers()
	if add {
		if containsFinalizer(finalizers, finalizerName) {
			return nil
		}

		// Add finalizer
		finalizers = append(finalizers, finalizerName)
	} else {
		if !containsFinalizer(finalizers, finalizerName) {
			return nil
		}

		// Remove finalizer
		finalizers = removeFinalizer(finalizers, finalizerName)
	}

	crdInstance.SetFinalizers(finalizers)

	_, err = s.dynamicClient.Resource(gvr).
		Update(context.TODO(), crdInstance, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to modify finalizer: %v", err)
	}

	return nil
}

// UpdateStatus updates the status of the resource
func (s *K8sStore[T]) UpdateStatus(id string, status map[string]interface{}) error {
	gvr := schema.GroupVersionResource{
		Group:    s.Crd.Spec.Group,
		Version:  s.Crd.Spec.Versions[0].Name,
		Resource: s.Crd.Spec.Names.Plural,
	}

	//FIXME doesnot look nice
	extensions := status["extensions"].(map[string]interface{})
	reconciliationInfo := extensions["reconciliationInfo"].(map[string]interface{})

	if reconciliationInfo[resource.ReconciliationState] != nil {
		if reconciliationInfo[resource.ReconciliationState] == resource.Deleted {
			InfoLogger.Printf("Deleting %v\n", id)
			err := s.ModifyFinalizer(id, false)
			if err == nil {
				var indexToRemove int
				for i, res := range s.currentResources {
					if res.GetID() == id {
						indexToRemove = i
						break
					}
				}
				s.currentResources = append(s.currentResources[:indexToRemove], s.currentResources[indexToRemove+1:]...)
			}
			return err
		}
	}
	name := id
	crdInstance, err := s.dynamicClient.Resource(gvr).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return errors.New("Error fetching CRD instance")
	}
	if err := unstructured.SetNestedField(crdInstance.Object, status, "status"); err != nil {
		return errors.New("Error setting status field")
	}
	updatedCrdInstance, err := s.dynamicClient.Resource(gvr).UpdateStatus(context.TODO(), crdInstance, metav1.UpdateOptions{})
	if err != nil {
		return errors.New("Error updating CRD status")
	}
	_ = updatedCrdInstance
	return nil
}
