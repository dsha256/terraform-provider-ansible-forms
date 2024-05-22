package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"

	"terraform-provider-ansible-forms/internal/restclient"
	"terraform-provider-ansible-forms/internal/utils"
)

// JobResourceModel describes the resource data model.
type JobResourceModel struct {
	ID          int64          `mapstructure:"id"`
	Start       string         `mapstructure:"start"`
	End         string         `mapstructure:"end"`
	User        string         `mapstructure:"user"`
	UserType    string         `mapstructure:"user_type"`
	JobType     string         `mapstructure:"job_type"`
	Extravars   map[string]any `mapstructure:"extravars"`
	Credentials map[string]any `mapstructure:"credentials"`
	Form        string         `mapstructure:"formName"`
	Status      string         `mapstructure:"status"`
	Message     string         `mapstructure:"message"`
	Target      string         `mapstructure:"target"`
	NoOfRecords int64          `mapstructure:"no_of_records"`
	Counter     int64          `mapstructure:"counter"`
	Output      string         `mapstructure:"output"`
	Data        string         `mapstructure:"data"`
	Approval    string         `mapstructure:"approval"`
}

// JobGetDataSourceModel ...
type JobGetDataSourceModel struct {
	ID          int64  `mapstructure:"id"`
	Start       string `mapstructure:"start"`
	End         string `mapstructure:"end"`
	User        string `mapstructure:"user"`
	UserType    string `mapstructure:"user_type"`
	JobType     string `mapstructure:"job_type"`
	Extravars   string `mapstructure:"extravars"`
	Credentials string `mapstructure:"credentials"`
	Form        string `mapstructure:"formName"`
	Status      string `mapstructure:"status"`
	Message     string `mapstructure:"message"`
	Target      string `mapstructure:"target"`
	NoOfRecords int64  `mapstructure:"no_of_records"`
	Counter     int64  `mapstructure:"counter"`
	Output      string `mapstructure:"output"`
	Data        string `mapstructure:"data"`
	Approval    string `mapstructure:"approval"`
}

// GetJobResponse describes GET job response.
type GetJobResponse struct {
	Status  string                `mapstructure:"status"`
	Message string                `mapstructure:"message"`
	Data    JobGetDataSourceModel `mapstructure:"data"`
}

// CreateJobResponse ...
type CreateJobResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Output struct {
			ID int64 `json:"id"`
		} `json:"output"`
		Error string `json:"error"`
	} `json:"data"`
}

// GetJobByID gets job info by id.
func GetJobByID(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string) (*JobGetDataSourceModel, error) {
	statusCode, response, err := r.GetNilOrOneRecord("job/"+id, nil, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading job info", fmt.Sprintf("error on GET job/: %s, statusCode %d", err, statusCode))
	}

	var apiResp *GetJobResponse
	if err = mapstructure.Decode(response, &apiResp); err != nil {
		return nil, errorHandler.MakeAndReportError("failed to decode response from GET job", fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("read job info: %#v", apiResp.Data))

	apiResp.Data.Status = apiResp.Status

	return &apiResp.Data, nil
}

// CreateJob creates a job.
func CreateJob(errorHandler *utils.ErrorHandler, r restclient.RestClient, data JobResourceModel) (*GetJobResponse, error) {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding job body", fmt.Sprintf("error on encoding POST job/ body: %s, body: %#v", err, data))
	}

	statusCode, response, err := r.CallCreateMethod("job/", nil, body) // Ansible Forms API does not allow querying.
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating job", fmt.Sprintf("error on POST job/: %s, statusCode %d", err, statusCode))
	}

	var resp *CreateJobResponse
	if err = mapstructure.Decode(response.Records[0], &resp); err != nil {
		return nil, errorHandler.MakeAndReportError("failed to decode response from POST job/", fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create svm source - udata: %#v", resp))

	return &GetJobResponse{Data: JobGetDataSourceModel{ID: resp.Data.Output.ID, Status: resp.Status}}, nil
}

// DeleteJobByID deletes a job by ID.
func DeleteJobByID(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string) error {
	statusCode, _, err := r.CallDeleteMethod("job/"+id, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting job info", fmt.Sprintf("error on DELETE job/: %s, statusCode %d", err, statusCode))
	}

	return nil
}
