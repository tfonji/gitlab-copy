package gitlab

// SecurityPolicyProjectLink holds the path of the security policy project
// linked to a group.
type SecurityPolicyProjectLink struct {
	FullPath string
}

// --- Read (GraphQL) ---

type securityPolicyProjectData struct {
	Namespace struct {
		SecurityPolicyProject *struct {
			FullPath string `json:"fullPath"`
		} `json:"securityPolicyProject"`
	} `json:"namespace"`
}

const securityPolicyProjectQuery = `
query($fullPath: ID!) {
  namespace(fullPath: $fullPath) {
    securityPolicyProject {
      fullPath
    }
  }
}`

// GetGroupSecurityPolicyProject returns the security policy project linked to
// a group, or nil if none is linked.
func (c *Client) GetGroupSecurityPolicyProject(groupPath string) (*SecurityPolicyProjectLink, error) {
	var data securityPolicyProjectData
	if err := c.graphql(securityPolicyProjectQuery, map[string]any{"fullPath": groupPath}, &data); err != nil {
		return nil, err
	}
	if data.Namespace.SecurityPolicyProject == nil {
		return nil, nil
	}
	return &SecurityPolicyProjectLink{
		FullPath: data.Namespace.SecurityPolicyProject.FullPath,
	}, nil
}

// --- Write (REST) ---

// LinkSecurityPolicyProject links a security policy project to a group.
// The project must already exist on the dest instance.
func (c *Client) LinkSecurityPolicyProject(groupPath string, projectFullPath string) error {
	return c.post("/groups/"+encodePath(groupPath)+"/security/policy_project", map[string]string{
		"full_path": projectFullPath,
	}, nil)
}
