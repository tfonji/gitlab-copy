package gitlab

import "fmt"

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

// LinkSecurityPolicyProject links a security policy project to a group
// using the securityPolicyProjectAssign GraphQL mutation.
func (c *Client) LinkSecurityPolicyProject(groupPath string, projectFullPath string) error {
	type assignData struct {
		SecurityPolicyProjectAssign struct {
			Errors []string `json:"errors"`
		} `json:"securityPolicyProjectAssign"`
	}

	const mutation = `
mutation($fullPath: ID!, $projectPath: ID!) {
  securityPolicyProjectAssign(input: { namespacePath: $fullPath, fullPath: $projectPath }) {
    errors
  }
}`

	var data assignData
	err := c.graphql(mutation, map[string]any{
		"fullPath":    groupPath,
		"projectPath": projectFullPath,
	}, &data)
	if err != nil {
		return err
	}
	if len(data.SecurityPolicyProjectAssign.Errors) > 0 {
		return fmt.Errorf("securityPolicyProjectAssign: %s", data.SecurityPolicyProjectAssign.Errors[0])
	}
	return nil
}
