package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/enrayga/omc-o2ims/internal/operator/resource"
)

// CRD is the data structure for CustomResourceDefinition.
//
// It is used to represent the CustomResourceDefinition resource in the
// JSONStore.
type ResourceInfo struct {
	Kind     string                 `json:"kind,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Spec     map[string]interface{} `json:"spec,omitempty"`
}

type ResourceStatus struct {
	Status map[string]interface{} `json:"status,omitempty"`
}

type JSONStore[T resource.Resource] struct {
	directory        string
	currentResources []T // List of resources that are currently present
}

const (
	resourceInfoSuffix      = "_info.json"
	resourceStatusSuffix    = "_status.json"
	resourceFinalizerSuffix = "_finalizer"
	resourceDeleteSuffix    = "_delete"
)

// GetResourceInfoName returns the name of the file for the given resource name.
func GetResourceInfoName(resourceName string) string {
	return resourceName + resourceInfoSuffix
}

// GetResourceStatusName returns the name of the file for the given resource name.
func GetResourceStatusName(resourceName string) string {
	return resourceName + resourceStatusSuffix
}

// GetResourceFinalizerName returns the name of the file for the given resource name.
func GetResourceFinalizerName(resourceName string) string {
	return resourceName + resourceFinalizerSuffix
}

// GetResourceDeleteName returns the name of the file for the given resource name.
func GetResourceDeleteName(resourceName string) string {
	return resourceName + resourceDeleteSuffix
}

// NewJSONStore creates a new instance of JSONStore.
func NewJSONStore[T resource.Resource](dir string, deleteContents bool) (*JSONStore[T], error) {

	if dir == "" {
		dir = os.TempDir() + "/JSONStore-" + fmt.Sprintf("%d", time.Now().UnixNano())
	}

	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else if !info.IsDir() {
		return nil, fmt.Errorf("%s is a file, not a directory", dir)
	}

	if deleteContents {
		files, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("can't read directory %s: %w", dir, err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			err = os.Remove(file.Name())
			if err != nil {
				return nil, fmt.Errorf("can't delete file %s: %w", file.Name(), err)
			}
		}
	}
	return &JSONStore[T]{directory: dir}, nil
}

func (s *JSONStore[T]) ReconcileList() error {

	const (
		New int = iota
		Current
		Deleting
	)
	//fmt.Printf("Reconciling List\n")

	files, err := os.ReadDir(s.directory)
	if err != nil {
		return fmt.Errorf("can't read directory %s: %w", s.directory, err)
	}

	for _, file := range files {
		var data interface{}
		//skip directories we dont care about
		if file.IsDir() {
			continue
		}

		state := New
		var existingRes T
		filename := file.Name()
		//fmt.Printf("filename: %s\n", filename)
		var name string
		var objectMap map[string]interface{}

		if !strings.HasSuffix(filename, resourceInfoSuffix) {
			//fmt.Printf("Skipping: %s!! as does not have _info.json\n", filename)
			continue
		}
		name = strings.TrimSuffix(filename, resourceInfoSuffix)

		file, err := os.Open(filepath.Join(s.directory, filename))
		if err != nil {
			fmt.Printf("Skipping: %s!! Opening failed  %v\n", filename, err)
			continue
		}
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&data); err != nil {
			fmt.Printf("Skipping: %s!! Decoding error!!  %v\n", filename, err)
			file.Close()
			continue
		}
		objectMap = data.(map[string]interface{})

		for _, res := range s.currentResources {
			if res.GetID() == name {
				existingRes = res
				state = Current
				//fmt.Printf("Reconciling resource: id %s\n", name)
				break
			}
		}
		file.Close()

		// Check if the file for the resource is marked for deletion
		deleteFlename := path.Join(s.directory, GetResourceDeleteName(name))
		_, err = os.Stat(deleteFlename)
		if err == nil {
			state = Deleting
		}

		// Check if user requested deletion

		switch state {
		case New:
			//fmt.Printf("New resource: %s\n", name)
			var zero T
			newResource := zero.GetNew()
			typedResource := newResource.(T)
			_ = typedResource.SetInitFields(name, objectMap)

			err = s.ModifyFinalizer(name, true)
			// If adding the finalizer fails, we'll need to decide how to proceed.
			// If we add the resource to the list and fail to add the finalizer,
			// we may have trouble during deletion. On the other hand, if we
			// ignore every polling and fail to add the finalizer, we'll also
			// have trouble. For now, we choose to proceed with the assumption
			// that the finalizer will eventually be added during next poll.
			if err != nil {
				fmt.Printf("Add Finalizer error: %v\n", err)
				continue
			}
			s.currentResources = append(s.currentResources, typedResource)

		case Current:
			//FIXME ME error checks etc
			//Apply changes if neded
			//fmt.Printf("Current resource: %s\n", filename)
			_, _ = existingRes.Compare(name, objectMap, true)
		case Deleting:
			//fmt.Printf("Deleting resource: %s\n", filename)
			// check if the finalizer is present
			finalizerFilename := path.Join(s.directory, GetResourceFinalizerName(name))
			infoFilename := path.Join(s.directory, GetResourceInfoName(name))
			statusFilename := path.Join(s.directory, GetResourceStatusName(name))
			_, err = os.Stat(finalizerFilename)
			if err != nil {
				if os.IsNotExist(err) {
					// finalizer file does not exist, delete info and _delete files
					// this is to handle the case if the finalizer was removed manually
					_ = os.Remove(infoFilename)
					_ = os.Remove(deleteFlename)
					_ = os.Remove(statusFilename)
					//FIXME TBD Add error handling
				}
			}
			// Check if resource has started delete action
			// if not initiate deletion
			deleting := existingRes.GetDeleteFlag()
			if !deleting {
				err = existingRes.SetDeleteFlag()
				if err != nil {
					fmt.Printf("Initiating delete: %v\n", err)
				}
			}
			continue
		}
	}
	return nil
}

// UpdateStatus implements Store.UpdateStatus
func (s *JSONStore[T]) UpdateStatus(id string, status map[string]interface{}) error {

	statusFilename := path.Join(s.directory, GetResourceStatusName(id))
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "temp_status_")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	// Write status to temporary file
	err = json.NewEncoder(tmpFile).Encode(status)
	if err != nil {
		return err
	}
	// Close temporary file
	err = tmpFile.Close()
	if err != nil {
		return err
	}

	// Rename temporary file to status file
	err = os.Rename(tmpFile.Name(), statusFilename)
	if err != nil {
		return err
	}

	return nil
}

// List implements Store.List
func (s *JSONStore[T]) List() ([]T, error) {

	var resources []T
	resources = append(resources, s.currentResources...)
	return resources, nil
}

// ModifyFinalizer adds or removes a finalizer from the specified resource
func (s *JSONStore[T]) ModifyFinalizer(name string, add bool) error {

	infoFilename := path.Join(s.directory, GetResourceInfoName(name))

	// check if the file exist
	_, err := os.Stat(infoFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("resource file %s not found", infoFilename)
		}
		return err
	}
	finalizerFilename := path.Join(s.directory, GetResourceFinalizerName(name))

	if add {
		// check if the finalizer is present
		finalizerData, err := os.ReadFile(finalizerFilename)
		_ = finalizerData
		if err == nil {
			//fmt.Printf("finalizer already present\n")
			// remove the finalizer
			_ = os.Remove(finalizerFilename)
			//FIXME Add error handling

			//fmt.Printf("finalizer removed\n")
		}

		_ = os.WriteFile(finalizerFilename, []byte{}, 0644)
		//FIXME Add error handling
		//fmt.Printf("finalizer added\n")
	} else {
		// if the finalizer file exists, delete it
		if !os.IsNotExist(err) {
			finalizerFilename := path.Join(s.directory, GetResourceFinalizerName(name))
			_ = os.Remove(finalizerFilename)
			//FIXME Add error handling
		}
	}
	return nil
}
