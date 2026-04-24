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

// --- Read (GraphQL) ---

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

// --- Write (GraphQL mutations) ---

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
