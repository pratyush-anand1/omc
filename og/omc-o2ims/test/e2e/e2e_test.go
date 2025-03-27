package e2e

import (
	"testing"
)

func TestRealK8sE2E(t *testing.T) {
	// 	crdDefinitions := []string{(`
	// # provisioning-request-crd.yaml
	// apiVersion: apiextensions.k8s.io/v1
	// kind: CustomResourceDefinition
	// metadata:
	//   name: provisioningrequests.o2ims.provisioning.oran.org
	// spec:
	//   group: o2ims.provisioning.oran.org
	//   names:
	//     kind: ProvisioningRequest
	//     listKind: ProvisioningRequestList
	//     plural: provisioningrequests
	//     singular: provisioningrequest
	//   scope: Cluster
	//   versions:
	//     - name: v1alpha1
	//       served: true
	//       storage: true
	//       schema:
	//         openAPIV3Schema:
	//           type: object
	//           properties:
	//             apiVersion:
	//               type: string
	//               description: |-
	//                 APIVersion defines the versioned schema of this representation of an object.
	//                 Servers should convert recognized schemas to the latest internal value, and
	//                 may reject unrecognized values.
	//                 The current apiVersion of this api is v1alpha1
	//             kind:
	//               type: string
	//               description: |-
	//                 Kind is a string value representing the REST resource this object represents.
	//                 Servers may infer this from the endpoint the client submits requests to.
	//                 Cannot be updated.
	//                 In CamelCase.
	//                 The kind value for this api is ProvisioningRequest
	//             metadata:
	//               type: object
	//               properties:
	//                 name:
	//                   type: string
	//                   description: |
	//                     The name of the ProvisioningRequest custom resource instance contains the provisioningItemId.
	//                     The provisioningItemId is the unique SMO provided identifier that the SMO will use to
	//                     identify all resources provisioned by this provisioning request in interactions
	//                     with the O-Cloud.
	//             spec:
	//               type: object
	//               properties:
	//                 name:
	//                   type: string
	//                   description: |
	//                     the name in this spec section is a human readable name intended for descriptive
	//                     purposes, this name is not required to be unique and does not identify a provisioning
	//                     request or any provisioned resources.
	//                 description:
	//                   type: string
	//                   description: |
	//                     A description of this provisioning request.
	//                 templateName:
	//                   type: string
	//                   description: |
	//                     templateName is the name of the template that the SMO wants to use to provision
	//                     resources
	//                 templateVersion:
	//                   type: string
	//                   description: |
	//                     templateVersion is the version of the template that the SMO wants to use to provision
	//                     resources. templateName and templateVersion together uniquely identify the template
	//                     instance that the SMO wants to use in the provisioning request.
	//                 templateParameters:
	//                   type: object
	//                   x-kubernetes-preserve-unknown-fields: true
	//                   description: |
	//                     templateParams carries the parameters required to provision resources using this template.
	//                     The type is object as actual parameters are defined by the template.
	//                     The template parameter schema itself is not defined here as it is template specific.
	//                     The themplate parameter schema must be published by the template provider so that FOCOM can
	//                     learn about required parameters and validate the same.
	//                     The template parameter schema language must be standardized by O-RAN.
	//               required:
	//                 - templateName
	//                 - templateVersion
	//                 - templateParameters
	//             status:
	//               type: object
	//               description: ProvisioningRequestStatus defines the observed state of ProvisioningRequest
	//               properties:
	//                 provisionedResources:
	//                   description: |
	//                     The resources that have been successfully provisioned as part of the provisioning process.
	//                   properties:
	//                     oCloudNodeClusterId:
	//                       description: |
	//                         The identifier of the provisioned oCloud NodeCluster.
	//                       type: string
	//                     oCloudInfrastructureResourceIds:
	//                       description: |
	//                         The list of provisioned infrastructure resource ids.
	//                       type: array
	//                       items:
	//                         type: string
	//                         description: |
	//                           The provisioned infrastructure resource id.
	//                   type: object
	//                 provisioningStatus:
	//                   properties:
	//                     provisioningUpdateTime:
	//                       description: |
	//                         The last update time of the provisioning status.
	//                       format: date-time
	//                       type: string
	//                     provisioningMessage:
	//                       description: |
	//                         The details about the current state of the provisioning process.
	//                       type: string
	//                     provisioningState:
	//                       description: The current state of the provisioning process.
	//                       enum:
	//                       - progressing
	//                       - fulfilled
	//                       - failed
	//                       - deleting
	//                       type: string
	//                   type: object
	//                 extensions:
	//                   description: |-
	//                     Extensions contain extra details about the resources and the configuration used for/by
	//                     the ProvisioningRequest.
	//                   type: object
	//                   x-kubernetes-preserve-unknown-fields: true
	//       subresources:
	//         status: {}`)}

	// 	namespace := "default"
	// 	kubeconfigPath := filepath.Join("testdata", "kubeconfig")

	// 	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		t.Skip("Skipping test: Cluster not available")
	// 		return
	// 	}
	// 	ctx := context.Background()
	// 	dynamicClient, err := dynamic.NewForConfig(config)

	// 	if err != nil {
	// 		fmt.Printf("Failed to create dynamic client: %v", err)
	// 		os.Exit(1)
	// 	} else {
	// 		fmt.Println("Successfully created dynamic client")
	// 	}

	// 	apiextensionsClient, err := apiextensionsclient.NewForConfig(config)
	// 	if err != nil {
	// 		fmt.Printf("Failed to create apiextensions client: %v", err)
	// 		os.Exit(1)
	// 	} else {
	// 		fmt.Println("Successfully created apiextensions client")
	// 	}

	// 	k8sStore, err := store.NewK8sStore[*res.ProvisioningRequest](
	// 		apiextensionsClient,
	// 		dynamicClient,
	// 		ctx, crdDefinitions[0], namespace)
	// 	if err != nil {
	// 		fmt.Printf("Failed to create k8sStore: %v", err)
	// 		os.Exit(1)
	// 	}

	// 	store := k8sStore
	// 	watcher := watcher.NewWatcherImpl[*res.ProvisioningRequest](store)
	// 	wg := sync.WaitGroup{}
	// 	wg.Add(1)

	// 	go func() {
	// 		defer wg.Done()
	// 		if err := watcher.StartWatching(ctx); err != nil {
	// 			fmt.Printf("Failed to start watching: %v", err)
	// 			os.Exit(1)
	// 		}
	// 	}()
	// 	wg.Wait()
	// 	select {} // wait forever

}

func performE2ETest() string {
	// Add your E2E test logic here
	// Example:
	// result := makeE2ERequest()
	// return result
	return "Dummy E2E Result"
}
