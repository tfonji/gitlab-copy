package gitlab

import "fmt"

// ComplianceFramework represents a group compliance framework.
type ComplianceFramework struct {
	ID                            string `json:"id"` // GID e.g. "gid://gitlab/ComplianceManagement::Framework/1"
	Name                          string `json:"name"`
	Description                   string `json:"description"`
	Color                         string `json:"color"`
	Default                       bool   `json:"default"`
	PipelineConfigurationFullPath string `json:"pipeline_configuration_full_path"`
	ProjectCount                  int    `json:"project_count"`
}

// ComplianceAssignment maps a project path to the framework names assigned to it.
type ComplianceAssignment struct {
	ProjectPath    string
	FrameworkNames []string
}

// --- Read frameworks (GraphQL) ---

type complianceFrameworksData struct {
	Namespace struct {
		ComplianceFrameworks struct {
			Nodes []struct {
				ID                            string `json:"id"`
				Name                          string `json:"name"`
				Description                   string `json:"description"`
				Color                         string `json:"color"`
				Default                       bool   `json:"default"`
				PipelineConfigurationFullPath string `json:"pipelineConfigurationFullPath"`
				Projects                      struct {
					Count int `json:"count"`
				} `json:"projects"`
			} `json:"nodes"`
		} `json:"complianceFrameworks"`
	} `json:"namespace"`
}

const complianceFrameworksQuery = `
query($fullPath: ID!) {
  namespace(fullPath: $fullPath) {
    complianceFrameworks {
      nodes {
        id
        name
        description
        color
        default
        pipelineConfigurationFullPath
        projects {
          count
        }
      }
    }
  }
}`

func (c *Client) GetGroupComplianceFrameworks(groupPath string) ([]ComplianceFramework, error) {
	var data complianceFrameworksData
	err := c.graphql(complianceFrameworksQuery, map[string]any{"fullPath": groupPath}, &data)
	if err != nil {
		return nil, err
	}
	frameworks := make([]ComplianceFramework, 0, len(data.Namespace.ComplianceFrameworks.Nodes))
	for _, n := range data.Namespace.ComplianceFrameworks.Nodes {
		frameworks = append(frameworks, ComplianceFramework{
			ID:                            n.ID,
			Name:                          n.Name,
			Description:                   n.Description,
			Color:                         n.Color,
			Default:                       n.Default,
			PipelineConfigurationFullPath: n.PipelineConfigurationFullPath,
			ProjectCount:                  n.Projects.Count,
		})
	}
	return frameworks, nil
}

// --- Read assignments (GraphQL) ---
// Returns all project paths under the group that have at least one compliance framework assigned.

type complianceAssignmentsData struct {
	Group struct {
		Projects struct {
			Nodes []struct {
				FullPath             string `json:"fullPath"`
				ComplianceFrameworks struct {
					Nodes []struct {
						Name string `json:"name"`
					} `json:"nodes"`
				} `json:"complianceFrameworks"`
			} `json:"nodes"`
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
		} `json:"projects"`
	} `json:"group"`
}

const complianceAssignmentsQuery = `
query($fullPath: ID!, $after: String) {
  group(fullPath: $fullPath) {
    projects(after: $after, includeSubgroups: true, first: 100) {
      nodes {
        fullPath
        complianceFrameworks {
          nodes {
            name
          }
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
}`

func (c *Client) GetGroupComplianceAssignments(groupPath string) ([]ComplianceAssignment, error) {
	var all []ComplianceAssignment
	var cursor *string

	for {
		vars := map[string]any{"fullPath": groupPath}
		if cursor != nil {
			vars["after"] = *cursor
		}

		var data complianceAssignmentsData
		if err := c.graphql(complianceAssignmentsQuery, vars, &data); err != nil {
			return nil, err
		}

		for _, proj := range data.Group.Projects.Nodes {
			if len(proj.ComplianceFrameworks.Nodes) == 0 {
				continue
			}
			names := make([]string, 0, len(proj.ComplianceFrameworks.Nodes))
			for _, fw := range proj.ComplianceFrameworks.Nodes {
				names = append(names, fw.Name)
			}
			all = append(all, ComplianceAssignment{
				ProjectPath:    proj.FullPath,
				FrameworkNames: names,
			})
		}

		if !data.Group.Projects.PageInfo.HasNextPage {
			break
		}
		c := data.Group.Projects.PageInfo.EndCursor
		cursor = &c
	}

	return all, nil
}

// --- Write frameworks (GraphQL mutations) ---

type createFrameworkData struct {
	CreateComplianceFramework struct {
		Framework struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"framework"`
		Errors []string `json:"errors"`
	} `json:"createComplianceFramework"`
}

const createComplianceFrameworkMutation = `
mutation($namespacePath: ID!, $params: ComplianceFrameworkInput!) {
  createComplianceFramework(input: { namespacePath: $namespacePath, params: $params }) {
    framework {
      id
      name
    }
    errors
  }
}`

func (c *Client) CreateComplianceFramework(groupPath string, f ComplianceFramework) (string, error) {
	params := map[string]any{
		"name":        f.Name,
		"description": f.Description,
		"color":       f.Color,
		"default":     f.Default,
	}
	if f.PipelineConfigurationFullPath != "" {
		params["pipelineConfigurationFullPath"] = f.PipelineConfigurationFullPath
	}

	var data createFrameworkData
	err := c.graphql(createComplianceFrameworkMutation, map[string]any{
		"namespacePath": groupPath,
		"params":        params,
	}, &data)
	if err != nil {
		return "", err
	}
	if len(data.CreateComplianceFramework.Errors) > 0 {
		return "", fmt.Errorf("createComplianceFramework: %s", data.CreateComplianceFramework.Errors[0])
	}
	return data.CreateComplianceFramework.Framework.ID, nil
}

// --- Delete framework (GraphQL mutation) ---

type destroyFrameworkData struct {
	DestroyComplianceFramework struct {
		Errors []string `json:"errors"`
	} `json:"destroyComplianceFramework"`
}

const unsetDefaultFrameworkMutation = `
mutation($id: ComplianceManagementFrameworkID!) {
  updateComplianceFramework(input: { id: $id, params: { default: false } }) {
    errors
  }
}`

const destroyComplianceFrameworkMutation = `
mutation($id: ComplianceManagementFrameworkID!) {
  destroyComplianceFramework(input: { id: $id }) {
    errors
  }
}`

func (c *Client) DeleteComplianceFramework(frameworkID string, isDefault bool) error {
	// Default frameworks cannot be deleted directly — must unset default first
	if isDefault {
		var unsetData struct {
			UpdateComplianceFramework struct {
				Errors []string `json:"errors"`
			} `json:"updateComplianceFramework"`
		}
		if err := c.graphql(unsetDefaultFrameworkMutation, map[string]any{"id": frameworkID}, &unsetData); err != nil {
			return fmt.Errorf("unsetting default on framework: %w", err)
		}
		if len(unsetData.UpdateComplianceFramework.Errors) > 0 {
			return fmt.Errorf("unsetting default on framework: %s", unsetData.UpdateComplianceFramework.Errors[0])
		}
	}

	var data destroyFrameworkData
	if err := c.graphql(destroyComplianceFrameworkMutation, map[string]any{"id": frameworkID}, &data); err != nil {
		return err
	}
	if len(data.DestroyComplianceFramework.Errors) > 0 {
		return fmt.Errorf("destroyComplianceFramework: %s", data.DestroyComplianceFramework.Errors[0])
	}
	return nil
}

const assignComplianceFrameworkMutation = `
mutation($projectId: ProjectID!, $frameworkIds: [ComplianceManagementFrameworkID!]!) {
  projectUpdateComplianceFrameworks(input: { projectId: $projectId, complianceFrameworkIds: $frameworkIds }) {
    errors
  }
}`

func (c *Client) AssignComplianceFrameworks(projectPath string, frameworkIDs []string) error {
	// Resolve project path to global ID
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
	if err := c.graphql(projectQuery, map[string]any{"fullPath": projectPath}, &pd); err != nil {
		return fmt.Errorf("resolving project ID: %w", err)
	}
	if pd.Project.ID == "" {
		return fmt.Errorf("project %q not found on dest", projectPath)
	}

	type assignData struct {
		ProjectUpdateComplianceFrameworks struct {
			Errors []string `json:"errors"`
		} `json:"projectUpdateComplianceFrameworks"`
	}
	var data assignData
	err := c.graphql(assignComplianceFrameworkMutation, map[string]any{
		"projectId":    pd.Project.ID,
		"frameworkIds": frameworkIDs,
	}, &data)
	if err != nil {
		return err
	}
	if len(data.ProjectUpdateComplianceFrameworks.Errors) > 0 {
		return fmt.Errorf("projectUpdateComplianceFrameworks: %s", data.ProjectUpdateComplianceFrameworks.Errors[0])
	}
	return nil
}
