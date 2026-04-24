package gitlab

import "net/url"

// Variable represents a CI/CD variable on a group or project.
type Variable struct {
	Key              string `json:"key"`
	Value            string `json:"value"`
	VariableType     string `json:"variable_type"`
	Protected        bool   `json:"protected"`
	Masked           bool   `json:"masked"`
	Hidden           bool   `json:"hidden"`
	Raw              bool   `json:"raw"`
	EnvironmentScope string `json:"environment_scope"`
	Description      string `json:"description"`
	AccessLevel      string `json:"access_level"`
}

// IsSensitive returns true for masked or hidden variables whose values
// cannot be read from the API. These are skipped during copy.
func (v *Variable) IsSensitive() bool {
	return v.Masked || v.Hidden
}

// VariableRequest is the write body for POST/PUT.
type VariableRequest struct {
	Key              string `json:"key"`
	Value            string `json:"value"`
	VariableType     string `json:"variable_type"`
	Protected        bool   `json:"protected"`
	Masked           bool   `json:"masked"`
	Raw              bool   `json:"raw"`
	EnvironmentScope string `json:"environment_scope"`
	Description      string `json:"description"`
}

// --- Read ---

func (c *Client) GetGroupVariables(groupPath string) ([]Variable, error) {
	var variables []Variable
	err := c.get("/groups/"+encodePath(groupPath)+"/variables", nil, &variables)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.IsForbidden() {
			return nil, nil
		}
		return nil, err
	}
	return variables, nil
}

// --- Write ---

// CreateGroupVariable creates a new CI/CD variable on the dest group.
func (c *Client) CreateGroupVariable(groupPath string, req VariableRequest) error {
	return c.post("/groups/"+encodePath(groupPath)+"/variables", req, nil)
}

// UpdateGroupVariable updates an existing variable on the dest group.
// The environment_scope filter is required when scope is not "*".
func (c *Client) UpdateGroupVariable(groupPath string, key string, scope string, req VariableRequest) error {
	params := url.Values{}
	if scope != "" && scope != "*" {
		params.Set("filter[environment_scope]", scope)
	}
	path := "/groups/" + encodePath(groupPath) + "/variables/" + encodePath(key)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}
	return c.put(path, req, nil)
}
