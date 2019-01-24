package awsssm

import (
	"bytes"
	"reflect"
	"testing"
)

type env struct {
	DatabasePassword string `mapstructure:"DB_PASSWORD"`
	DatabaseHost     string `mapstructure:"DB_HOST"`
	NotMappedField   string `mapstructure:"NOT_THERE"`
}

func TestParameters_GetValueByName(t *testing.T) {
	tests := []struct {
		name          string
		basePath      string
		parameters    map[string]*Parameter
		paramName     string
		expectedValue string
	}{
		{
			name:          "Success Get By Name",
			basePath:      "/my-service/dev/",
			parameters:    getParametersMap(),
			paramName:     "DB_PASSWORD",
			expectedValue: "something-secure",
		},
		{
			name:          "Success Get By Name That Doesn't Exist",
			basePath:      "/my-service/dev/",
			parameters:    getParametersMap(),
			paramName:     "NOT_EXISTING_PARAMETER",
			expectedValue: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parameter := NewParameters(test.basePath, test.parameters)
			value := parameter.GetValueByName(test.paramName)
			if value != test.expectedValue {
				t.Errorf(`Unexpected value: got %s, expected %s`, value, test.expectedValue)
			}
		})
	}
}

func TestParameters_GetValueByFullPath(t *testing.T) {
	tests := []struct {
		name          string
		basePath      string
		parameters    map[string]*Parameter
		paramName     string
		expectedValue string
	}{
		{
			name:          "Success Get By Name",
			basePath:      "/my-service/dev/",
			parameters:    getParametersMap(),
			paramName:     "/my-service/dev/DB_PASSWORD",
			expectedValue: "something-secure",
		},
		{
			name:          "Success Get By Name That Doesn't Exist",
			basePath:      "/my-service/dev/",
			parameters:    getParametersMap(),
			paramName:     "/my-service/dev/NOT_EXISTING_PARAMETER",
			expectedValue: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parameter := NewParameters(test.basePath, test.parameters)
			value := parameter.GetValueByFullPath(test.paramName)
			if value != test.expectedValue {
				t.Errorf(`Unexpected value: got %s, expected %s`, value, test.expectedValue)
			}
		})
	}
}

func TestParameters_Decode(t *testing.T) {
	tests := []struct {
		name              string
		basePath          string
		parameters        map[string]*Parameter
		expectedError     error
		expectedEnvStruct *env
	}{
		{
			name:       "Success Decode",
			basePath:   "/my-service/dev/",
			parameters: getParametersMap(),
			expectedEnvStruct: &env{
				DatabasePassword: "something-secure",
				DatabaseHost:     "rds.something.aws.com",
				NotMappedField:   "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := new(env)
			parameter := NewParameters(test.basePath, test.parameters)
			err := parameter.Decode(e)
			if err != test.expectedError {
				t.Errorf(`Unexpected error: got %s, expected %s`, err, test.expectedError)
			}
			if !reflect.DeepEqual(e, test.expectedEnvStruct) {
				t.Errorf(`Unexpected value: got %s, expected %s`, e, test.expectedEnvStruct)
			}
		})
	}
}

func TestParameters_JsonMarshal_Success(t *testing.T) {

	jsonByteParams := []byte(`{"DB_HOST":"rds.something.aws.com","DB_PASSWORD":"something-secure"}`)
	basePath := "/my-service/dev/"
	parameters := getParametersMap()

	parameter := NewParameters(basePath, parameters)
	jsonBytes, err := parameter.JSONMarshal()
	if err != nil {
		t.Errorf(`Unexpected error: %s`, err)
	}
	if bytes.Compare(jsonBytes, jsonByteParams) != 0 {
		t.Errorf(`Unexpected value: got %s, expected %s`, jsonBytes, jsonByteParams)
	}

}
func getParametersMap() map[string]*Parameter {
	return map[string]*Parameter{
		"/my-service/dev/DB_PASSWORD": {ssmParameter: param1},
		"/my-service/dev/DB_HOST":     {ssmParameter: param2},
	}
}
