package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/enrayga/omc-o2ims/internal/operator/resource"
)

func TestNewJSONStore(t *testing.T) {
	existing_dir := "/tmp/JSON_NEWSTORE_EXISTING_DIR"

	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    *JSONStore[*(resource.MockResource)]
		wantErr bool
	}{
		{
			name:    "empty directory",
			args:    args{dir: ""},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "existing directory",
			args:    args{dir: existing_dir},
			want:    &JSONStore[*resource.MockResource]{directory: existing_dir},
			wantErr: false,
		},
		{
			name:    "invalid directory",
			args:    args{dir: "/path/to/nonexistent/directory"},
			want:    nil,
			wantErr: true,
		},
	}

	if err := os.MkdirAll(existing_dir, os.ModePerm); err != nil {
		t.Fatalf("Unable to create temp directory %v: %v", existing_dir, err)
	}

	defer func() {
		for _, tt := range tests {
			if tt.wantErr {
				continue
			}
			if err := os.RemoveAll(tt.args.dir); err != nil {
				fmt.Printf("unable to delete directory %v: %v\n", tt.args.dir, err)
			}

			fmt.Printf("Deleted directory %v:\n", tt.args.dir)

		}

	}()

	for i, tt := range tests {
		_ = i
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewJSONStore[*resource.MockResource](tt.args.dir, true)

			if !tt.wantErr && (err != nil) {
				t.Errorf("NewJSONStore()  dir = %v, error = %v, wantErr %v", tt.args.dir, err, tt.wantErr)
				return
			}
			if (err == nil) && (tt.want != nil) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewJSONStore() = %v, want %v", got, tt.want)
			}

			if tt.wantErr != true {
				tests[i].args.dir = got.directory
			}
		})
	}
}

func TestJSONStore_ReconcileList(t *testing.T) {

	type args struct {
		dir string
	}
	tests := []struct {
		name       string
		args       args
		resNames   []string
		resContent []string
		want       *JSONStore[*resource.MockResource]
		wantErr    bool
	}{
		{
			name:     "Test when new resources are created",
			args:     args{dir: ""},
			want:     nil,
			resNames: []string{"crdi-1", "crdi-2", "crdi-3"},
			resContent: []string{
				`{
					"id": "crdi-1",
					"name": "crdi-1",
					"value": "crdi-1-content"
				}`,
				`{
					"id": "crdi-2",
					"name": "crdi-2",
					"value": "crdi-2-content"
				}`,
				`{
					"id": "crdi-3",
					"name": "crdi-3",
					"value": "crdi-3-content"
				}`,
			},
			wantErr: false,
		},
	}

	defer func() {
		for _, tt := range tests {
			if err := os.RemoveAll(tt.args.dir); err != nil {
				fmt.Printf("unable to delete directory %v: %v\n", tt.args.dir, err)
			}

			fmt.Printf("Deleted directory %v:\n", tt.args.dir)

		}
	}()

	//Happy Path 3 resources should get added
	for i, _ := range tests {

		t.Run(tests[i].name, func(t *testing.T) {
			got, err := NewJSONStore[*resource.MockResource](tests[i].args.dir, true)

			if err != nil {
				t.Errorf("NewJSONStore() error = %v, wantErr %v", err, tests[i].wantErr)
				return
			}
			tests[i].args.dir = got.directory
			for j, name := range tests[i].resNames {

				file_name := GetResourceInfoName(name)
				file_name = fmt.Sprintf("%s/%s", tests[i].args.dir, file_name)

				file, err := os.Create(file_name)
				file.WriteString(tests[i].resContent[j])

				if err != nil {
					t.Errorf("Unable to create test file %v: %v", name, err)
					return
				}
				//fmt.Printf("Created test file %v\n", file.Name())
				file.Close()
			}

			err = got.ReconcileList()
			if (err != nil) != tests[i].wantErr {
				t.Errorf("ReconcileList() error = %v, wantErr %v", err, tests[i].wantErr)
			}

			lst, err := got.List()
			if err != nil {
				t.Errorf("List() error = %v", err)
			}
			if len(lst) != len(tests[i].resNames) {
				t.Errorf("List() expected length %v, got %v", len(tests[i].resNames), len(lst))
			}

			// Bad files should not get adde due to bad content
			file_name := GetResourceInfoName("bad_file")
			file_name = fmt.Sprintf("%s/%s", tests[i].args.dir, file_name)

			file, err := os.Create(file_name)
			file.WriteString("{bad json file}")

			if err != nil {
				t.Errorf("Unable to create test file %v: %v", file_name, err)
				return
			}
			//fmt.Printf("Created test file %v\n", file.Name())
			file.Close()

			// Directory instead of file should not get added
			dir_name := GetResourceInfoName("empty_dir")
			dir_name = fmt.Sprintf("%s/%s", tests[i].args.dir, dir_name)

			err = os.Mkdir(dir_name, os.ModePerm)
			if err != nil {
				t.Errorf("Unable to create test directory %v: %v", dir_name, err)
				return
			}
			//fmt.Printf("Created test directory %v\n", dir_name)

			file_name = GetResourceInfoName("unopenable_file")
			file_name = fmt.Sprintf("%s/%s", tests[i].args.dir, file_name)

			file, err = os.Create(file_name)
			file.Close()

			// Unreadable file should not get added
			err = os.Chmod(file_name, 0000)
			if err != nil {
				t.Errorf("Unable to change test file permission %v: %v", file_name, err)
				return
			}
			//fmt.Printf("Changed test file permission %v\n", file_name)

			// Call ReconcileList again, the "bad_file" resource should not be added to the list
			err = got.ReconcileList()
			if (err != nil) != tests[i].wantErr {
				t.Errorf("ReconcileList() error = %v, wantErr %v", err, tests[i].wantErr)
			}

			lst, err = got.List()
			if err != nil {
				t.Errorf("List() error = %v", err)
			}
			if len(lst) != len(tests[i].resNames) {
				t.Errorf("List() expected length %v, got %v", len(tests[i].resNames), len(lst))
			}

			// TBD add a test to check if finalizer creation has failed
			//Check when finaliser is created  before hand we should fail adding to the list

		})
	}
}

// Tests to simulate resource deletion
func TestReconcileListDeleteResource(t *testing.T) {
	dir := "/tmp/JSON_DELETE_RESOURCE_TEST"

	defer os.RemoveAll(dir)

	got, err := NewJSONStore[*resource.MockResource](dir, true)

	if err != nil {
		t.Errorf("NewJSONStore() error = %v", err)
		return
	}

	res1_name := "res-1"

	res1_content := `{
		"id": "res1_name",
		"name": "res1_name",
		"value": "res1_name"
	}`

	res1_file_name := GetResourceInfoName(res1_name)
	res1_file_name = fmt.Sprintf("%s/%s", dir, res1_file_name)
	file, err := os.Create(res1_file_name)
	file.WriteString(res1_content)

	err = got.ReconcileList()
	if err != nil {
		t.Errorf("ReconcileList() error = %v", err)
		return
	}

	lst, err := got.List()
	if err != nil {
		t.Errorf("List() error = %v", err)
		return
	}

	if len(lst) != 1 {
		t.Errorf("List() expected length of 1, got %v", len(lst))
		return
	}

	if lst[0].GetID() != res1_name {
		t.Errorf("List() expected id of %s, got %s", res1_name, lst[0].GetID())
	}

	// check if finalizer is present
	finalizerFilename := path.Join(dir, GetResourceFinalizerName(res1_name))
	_, err = os.Stat(finalizerFilename)
	if os.IsNotExist(err) {
		t.Errorf("finalizer file %s not found", finalizerFilename)
		return
	}

	// add a delete file to simulate user requested for delete
	deleteFileName := path.Join(dir, GetResourceDeleteName(res1_name))
	file, err = os.Create(deleteFileName)
	if err != nil {
		t.Errorf("Create() error = %v", err)
		return
	}
	file.Close()

	err = got.ReconcileList()
	if err != nil {
		t.Errorf("ReconcileList() error = %v", err)
		return
	}

	// Delete finalizer to simulate applcation is allowing deletion to preoceed
	err = got.ModifyFinalizer(res1_name, false)
	if err != nil {
		t.Errorf("ModifyFinalizer() error = %v", err)
		return
	}

	err = got.ReconcileList()
	if err != nil {
		t.Errorf("ReconcileList() error = %v", err)
		return
	}

	// check for exisitence of any files having res1_name in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Errorf("ReadDir() error = %v", err)
		return
	}

	for _, file := range files {
		if strings.Contains(file.Name(), res1_name) {
			t.Errorf("ReconcileList() expected file name %s to be deleted, but found %s", res1_name, file.Name())
			return
		}
	}
}

func TestJSONStore_UpdateStatus(t *testing.T) {
	dir, err := os.MkdirTemp("", "store_test")
	if err != nil {
		t.Errorf("TempDir() error = %v", err)
		return
	}
	defer os.RemoveAll(dir)

	store, err := NewJSONStore[*resource.MockResource](dir, true)
	if err != nil {
		t.Errorf("NewJSONStore() error = %v", err)
		return
	}

	res1_name := "res-1"

	status := map[string]interface{}{
		"key": "value",
	}
	err = store.UpdateStatus(res1_name, status)
	if err != nil {
		t.Errorf("getStatus() error = %v", err)
		return
	}

	err = store.UpdateStatus(res1_name, status)
	if err != nil {
		t.Errorf("UpdateStatus() error = %v", err)
		return
	}

	filePath := path.Join(dir, GetResourceStatusName(res1_name))
	file, err := os.Open(filePath)
	if err != nil {
		t.Errorf("Open() error = %v", err)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		t.Errorf("ReadAll() error = %v", err)
		return
	}

	var fileStatus map[string]interface{}
	err = json.Unmarshal(data, &fileStatus)
	if err != nil {
		t.Errorf("Unmarshal() error = %v", err)
		return
	}

	if !reflect.DeepEqual(status, fileStatus) {
		t.Errorf("status not matching. Expected: %v Got: %v", status, fileStatus)
	}
}
