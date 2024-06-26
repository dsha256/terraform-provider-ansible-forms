package restclient

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
)

// RestError maps the REST error structure
type RestError struct {
	Code    string
	Message string
	Target  string
}

// RestResponse to return a list of records (can be empty) and/or errors.
type RestResponse struct {
	NumRecords int `mapstructure:"num_records"`
	Records    []map[string]any
	RestError  RestError `mapstructure:"error"`
	StatusCode int
	HTTPError  string
	ErrorType  string
	Job        map[string]any
	Jobs       []map[string]any
}

// unmarshalResponse converts the REST response into a structure with a list of 0 or more records.
// We're doing it in two phases:
// 1. Unmarshall to intermediate structure, as records may or may not present.
// 2. Adjust intermediate structure, and decode to final structure.
func (r *RestClient) unmarshalResponse(statusCode int, responseJSON []byte, httpClientErr error) (int, RestResponse, error) {
	emptyResponse := RestResponse{
		NumRecords: 0,
		Records:    []map[string]any{},
		RestError:  RestError{},
		StatusCode: statusCode,
		HTTPError:  "",
		ErrorType:  "",
	}
	if httpClientErr != nil {
		emptyResponse.HTTPError = httpClientErr.Error()
		emptyResponse.ErrorType = "http"
		return statusCode, emptyResponse, httpClientErr
	}

	// We don't know which fields are present or not, and fields may not be in a record, so just use any
	var dataMap map[string]any
	if err := json.Unmarshal(responseJSON, &dataMap); err != nil {
		tflog.Error(r.ctx, fmt.Sprintf("unable to unmarshall response, this may be expected when statusCode %d >= 300, unmarshall error=%s, response=%#v", statusCode, err, responseJSON))
		emptyResponse.ErrorType = "bad_response_decode_json"
		return statusCode, emptyResponse, err
	}
	tflog.Debug(r.ctx, fmt.Sprintf("dataMap %#v", dataMap))

	// The returned REST response may or may not contain records.
	// If records is not present, the contents will show in Other.
	type restStagedResponse struct {
		NumRecords int `mapstructure:"num_records"`
		Records    []map[string]any
		Error      RestError
		Job        map[string]any
		Jobs       []map[string]any
		Other      map[string]any `mapstructure:",remain"`
	}

	var rawResponse restStagedResponse
	var metadata mapstructure.Metadata
	if err := mapstructure.DecodeMetadata(dataMap, &rawResponse, &metadata); err != nil {
		tflog.Error(r.ctx, fmt.Sprintf("unable to format raw response, this may be expected when statusCode %d >= 300, unmarshall error=%s, response=%#v", statusCode, err, dataMap))
		emptyResponse.ErrorType = "bad_response_decode_interface"
		return statusCode, emptyResponse, err
	}

	tflog.Debug(r.ctx, fmt.Sprintf("rawResponse %#v, metadata %#v", rawResponse, metadata))

	// If Other is present, add it to records.
	// But ignore it if we already have some records.
	// Other will always have 1 element called _link, so only do this if Other has more than 1 element
	// Examples:
	// {NumRecords:0 Records:[] Error:{Code: Message: Target:} Job:map[] Jobs:[] Other:map[_links:map[self:map[href:/api/cluster/schedules?fields=name%2Cuuid%2Ccron%2Cinterval%2Ctype%2Cscope&name=mytest]]]}
	// {NumRecords:0 Records:[] Error:{Code: Message: Target:} Job:map[] Jobs:[] Other:map[_links:map[self:map[href:/api/cluster]] certificate:map[_links:map[self:map[href:/api/security/certificates/2f632ea7-92cd-11ed-8f2b-005056b3357c]] uuid:2f632ea7-92cd-11ed-8f2b-005056b3357c] metric:map[duration:PT15S iops:map[other:0 read:0 total:0 write:0] latency:map[other:0 read:0 total:0 write:0] status:ok throughput:map[other:0 read:0 total:0 write:0] timestamp:2023-03-16T18:36:30Z] name:laurentncluster-2 peering_policy:map[authentication_required:true encryption_required:false minimum_passphrase_length:8] san_optimized:false statistics:map[iops_raw:map[other:0 read:0 total:0 write:0] latency_raw:map[other:0 read:0 total:0 write:0] status:ok throughput_raw:map[other:0 read:0 total:0 write:0] timestamp:2023-03-16T18:36:31Z] timezone:map[name:Etc/UTC] uuid:2115008a-92cd-11ed-8f2b-005056b3357c version:map[full:NetApp Release Metropolitan__9.11.1: Sat Dec 10 19:08:07 UTC 2022 generation:9 major:11 minor:1]]}
	if rawResponse.NumRecords == 0 && len(rawResponse.Records) == 0 && len(rawResponse.Other) > 1 {
		rawResponse.NumRecords = 1
		rawResponse.Records = append(rawResponse.Records, rawResponse.Other)
	}

	var finalResponse RestResponse
	if err := mapstructure.DecodeMetadata(rawResponse, &finalResponse, &metadata); err != nil {
		tflog.Error(r.ctx, fmt.Sprintf("unable to format final response - statusCode %d, http err=%#v, decode error=%s, response=%#v", statusCode, httpClientErr, err, rawResponse))
		emptyResponse.ErrorType = "bad_response_decode_raw"
		return statusCode, emptyResponse, err
	}

	// If we reached this point, the only possible errors are a bad HTTP status code and/or a REST error encoded in the paybload
	finalResponse.StatusCode = statusCode
	finalResponse, err := r.checkRestErrors(statusCode, finalResponse)
	tflog.Debug(r.ctx, fmt.Sprintf("finalResponse %#v, metadata %#v", finalResponse, metadata))

	return statusCode, finalResponse, err
}

// check for statusCode and RestError
func (r *RestClient) checkRestErrors(statusCode int, response RestResponse) (RestResponse, error) {
	var err error
	if response.RestError.Code != "0" && response.RestError.Code != "" {
		response.ErrorType = "rest_error"
		err = fmt.Errorf("REST reported error %#v, statusCode: %d", response.RestError, statusCode)
	} else if err = r.checkStatusCode(statusCode); err != nil {
		response.ErrorType = "statuscode_error"
	}
	if err != nil {
		tflog.Error(r.ctx, fmt.Sprintf("checkRestError: %s, statusCode %d, response: %#v", err, statusCode, response))
	}

	return response, err
}

// checkStatusCode checks and validates the statusCode
func (r *RestClient) checkStatusCode(statusCode int) error {
	if statusCode >= 300 || statusCode < 200 {
		return fmt.Errorf("statusCode indicates error, without details: %d", statusCode)
	}

	return nil
}
