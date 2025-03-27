package resource

type MockResource struct {
	Id              string
	Name            string
	Status          string
	Type            string
	Spec            string
	Description     string
	TemplateName    string
	TemplateVersion string
	TemplateParams  map[string]interface{}
}

func (s *MockResource) GetID() string {
	// FIXME value by pointer
	return s.Id
}

func (s *MockResource) GetName() string {
	// Implement the logic to return the name of the resource
	// ...
	return s.Name
}

func (s *MockResource) GetType() (map[string]string, error) {
	// Implement the logic to return the type of the resource
	// ...
	return nil, nil
}

func (s *MockResource) GetMetadata() (map[string]interface{}, error) {
	// Implement the logic to return the metadata of the resource
	// ...
	return nil, nil
}

func (s *MockResource) GetSpec() (map[string]interface{}, error) {
	// Implement the logic to return the desired spec of the resource
	// ...
	return nil, nil
}

func (s *MockResource) GetStatus() (map[string]interface{}, error) {
	// Implement the logic to return the desired spec of the resource
	// ...
	return nil, nil
}

func (s *MockResource) Reconcile() error {
	return nil
}

func (r *MockResource) UpdateStatus(currentStatus Status) error {
	// Implement the UpdateStatus method
	return nil
}

func (r *MockResource) HandleEvent(event NotifierEvent) {
	// Implement the HandleEvent method
}

func (r *MockResource) Compare(name string, fields map[string]interface{}, apply bool) (bool, error) {
	// Implement the Compare method
	return false, nil
}

func (r *MockResource) GetDeleteFlag() bool {
	// Implement the GetDeleteFlag method
	return false
}

func (r *MockResource) SetDeleteFlag() error {
	// Implement the SetDeleteFlag method
	return nil
}

func (r *MockResource) GetNew() interface{} {
	res := &MockResource{}
	return res
}

func (r *MockResource) SetInitFields(name string, fields map[string]interface{}) error {
	r.Id = name
	r.Name = name
	r.Status = ""
	r.Type = ""
	r.Spec = ""
	r.Description = ""
	r.TemplateName = ""
	r.TemplateVersion = ""
	r.TemplateParams = map[string]interface{}{}
	return nil
}
