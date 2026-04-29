package gitlab

import "fmt"

// DeployKey represents a project deploy key.
type DeployKey struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at"`
	Key       string `json:"key"` // public key content
	CanPush   bool   `json:"can_push"`
}

// DeployKeyRequest is the write body for POST /projects/:id/deploy_keys.
type DeployKeyRequest struct {
	Title   string `json:"title"`
	Key     string `json:"key"`
	CanPush bool   `json:"can_push"`
}

// --- Read ---

func (c *Client) GetProjectDeployKeys(projectID int) ([]DeployKey, error) {
	var keys []DeployKey
	err := c.get(fmt.Sprintf("/projects/%d/deploy_keys", projectID), nil, &keys)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.IsForbidden() {
			return nil, nil
		}
		return nil, err
	}
	return keys, nil
}

// --- Write ---

// CreateProjectDeployKey creates a deploy key on the dest project.
// If the key already exists globally on the dest instance (same public key
// content), GitLab returns 422. In that case the caller should report a
// failure with a message to enable the key manually via the UI.
func (c *Client) CreateProjectDeployKey(projectID int, req DeployKeyRequest) error {
	return c.post(fmt.Sprintf("/projects/%d/deploy_keys", projectID), req, nil)
}
