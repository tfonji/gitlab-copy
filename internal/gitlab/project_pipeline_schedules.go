package gitlab

import "fmt"

// PipelineSchedule represents a project pipeline schedule.
type PipelineSchedule struct {
	ID           int    `json:"id"`
	Description  string `json:"description"`
	Ref          string `json:"ref"`
	Cron         string `json:"cron"`
	CronTimezone string `json:"cron_timezone"`
	Active       bool   `json:"active"`
}

// PipelineScheduleVariable is a variable attached to a pipeline schedule.
type PipelineScheduleVariable struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	VariableType string `json:"variable_type"`
}

// PipelineScheduleRequest is the write body for POST/PUT.
type PipelineScheduleRequest struct {
	Description  string `json:"description"`
	Ref          string `json:"ref"`
	Cron         string `json:"cron"`
	CronTimezone string `json:"cron_timezone"`
	Active       bool   `json:"active"`
}

// PipelineScheduleVariableRequest is the write body for schedule variable POST.
type PipelineScheduleVariableRequest struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	VariableType string `json:"variable_type"`
}

// --- Read ---

func (c *Client) GetProjectPipelineSchedules(projectPath string) ([]PipelineSchedule, error) {
	var schedules []PipelineSchedule
	err := c.get("/projects/"+encodePath(projectPath)+"/pipeline_schedules", nil, &schedules)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	return schedules, nil
}

func (c *Client) GetPipelineScheduleVariables(projectPath string, scheduleID int) ([]PipelineScheduleVariable, error) {
	var vars []PipelineScheduleVariable
	path := fmt.Sprintf("/projects/%s/pipeline_schedules/%d/variables", encodePath(projectPath), scheduleID)
	err := c.get(path, nil, &vars)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	return vars, nil
}

// --- Write ---

func (c *Client) CreateProjectPipelineSchedule(projectPath string, req PipelineScheduleRequest) (int, error) {
	var resp PipelineSchedule
	if err := c.post("/projects/"+encodePath(projectPath)+"/pipeline_schedules", req, &resp); err != nil {
		return 0, err
	}
	return resp.ID, nil
}

func (c *Client) CreatePipelineScheduleVariable(projectPath string, scheduleID int, req PipelineScheduleVariableRequest) error {
	path := fmt.Sprintf("/projects/%s/pipeline_schedules/%d/variables", encodePath(projectPath), scheduleID)
	return c.post(path, req, nil)
}
