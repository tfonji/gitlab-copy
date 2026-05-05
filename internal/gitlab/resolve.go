package gitlab

import (
	"fmt"
	"net/url"
)

// ProjectSearchResult is a minimal project from the search API.
type ProjectSearchResult struct {
	ID                int    `json:"id"`
	PathWithNamespace string `json:"path_with_namespace"`
	Namespace         struct {
		FullPath string `json:"full_path"`
	} `json:"namespace"`
	Topics []string `json:"topics"`
}

// FindProjectsByTopic returns all projects on this instance that have the
// given topic tag. Paginates automatically.
func (c *Client) FindProjectsByTopic(topic string) ([]ProjectSearchResult, error) {
	var all []ProjectSearchResult
	page := 1

	for {
		params := url.Values{}
		params.Set("topic", topic)
		params.Set("per_page", "100")
		params.Set("page", fmt.Sprintf("%d", page))
		params.Set("archived", "false")

		var batch []ProjectSearchResult
		if err := c.get("/projects", params, &batch); err != nil {
			return nil, fmt.Errorf("searching projects by topic %q: %w", topic, err)
		}
		if len(batch) == 0 {
			break
		}
		all = append(all, batch...)
		if len(batch) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// GroupExists checks whether a group exists on this instance.
func (c *Client) GroupExists(groupPath string) (bool, error) {
	var g struct {
		ID int `json:"id"`
	}
	err := c.get("/groups/"+encodePath(groupPath), nil, &g)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return false, nil
		}
		return false, err
	}
	return g.ID > 0, nil
}
