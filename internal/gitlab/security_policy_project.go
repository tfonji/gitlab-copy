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
	// Step 1 — resolve the project full path to its GraphQL GID
	type projectData struct {
		Project struct {
			ID string `json:"id"`
		} `json:"project"`
	}
	const projectQuery = `
query($fullPath: ID!) {
  project(fullPath: $fullPath) {
    id
  }
}`
	var pd projectData
	if err := c.graphql(projectQuery, map[string]any{"fullPath": projectFullPath}, &pd); err != nil {
		return fmt.Errorf("resolving policy project ID: %w", err)
	}
	if pd.Project.ID == "" {
		return fmt.Errorf("policy project %q not found", projectFullPath)
	}

	// Step 2 — assign using the GID
	type assignData struct {
		SecurityPolicyProjectAssign struct {
			Errors []string `json:"errors"`
		} `json:"securityPolicyProjectAssign"`
	}
	const mutation = `
mutation($fullPath: String!, $securityPolicyProjectId: ProjectID!) {
  securityPolicyProjectAssign(input: { fullPath: $fullPath, securityPolicyProjectId: $securityPolicyProjectId }) {
    errors
  }
}`
	var data assignData
	err := c.graphql(mutation, map[string]any{
		"fullPath":                groupPath,
		"securityPolicyProjectId": pd.Project.ID,
	}, &data)
	if err != nil {
		return err
	}
	if len(data.SecurityPolicyProjectAssign.Errors) > 0 {
		return fmt.Errorf("securityPolicyProjectAssign: %s", data.SecurityPolicyProjectAssign.Errors[0])
	}
	return nil
}
