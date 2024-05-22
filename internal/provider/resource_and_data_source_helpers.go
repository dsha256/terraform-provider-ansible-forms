package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"terraform-provider-ansible-forms/internal/restclient"
	"terraform-provider-ansible-forms/internal/utils"
)

type resourceOrDataSourceConfig struct {
	client         *restclient.RestClient
	providerConfig Config
	name           string
}

// getRestClient will use existing client config.client or create one if it's not set
func getRestClient(errorHandler *utils.ErrorHandler, config resourceOrDataSourceConfig, cxProfileName types.String) (*restclient.RestClient, error) {

	if config.client == nil {
		client, err := config.providerConfig.NewClient(errorHandler, cxProfileName.ValueString(), config.name)
		if err != nil {
			return nil, err
		}
		config.client = client
	}
	return config.client, nil
}

// func flattenTypesInt64List(clist []int64) interface{} {
func flattenTypesInt64List(clist []int64) []types.Int64 {
	if len(clist) == 0 {
		return nil
	}
	cronUnits := make([]types.Int64, len(clist))
	for index, record := range clist {
		cronUnits[index] = types.Int64Value(record)
	}

	return cronUnits
}

// func flattenTypesStringList(terraformStringsList []string) interface{} {
func flattenTypesStringList(terraformStringsList []string) []types.String {
	if len(terraformStringsList) == 0 {
		return nil
	}
	stringsList := make([]types.String, len(terraformStringsList))
	for index, record := range terraformStringsList {
		stringsList[index] = types.StringValue(record)
	}

	return stringsList
}

// jsonStringToMapValue converts JSON string to basetypes.MapType.
func jsonStringToMapValue(ctx context.Context, diags *diag.Diagnostics, str string) basetypes.MapValue {
	var credentialsMap map[string]interface{}
	err := json.Unmarshal([]byte(str), &credentialsMap)
	if err != nil {
		diags.AddError("error unmarshalling JSON string", err.Error())
		return basetypes.MapValue{}
	}

	for k, v := range credentialsMap {
		switch v.(type) {
		case string:
		default:
			delete(credentialsMap, k)
		}
	}

	m, d := types.MapValueFrom(ctx, types.StringType, credentialsMap)
	diags.Append(d...)

	return m
}
