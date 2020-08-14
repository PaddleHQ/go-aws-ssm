package awsssm

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"reflect"
	"testing"
)

var param1 = new(ssm.Parameter).
	SetName("/my-service/dev/DB_PASSWORD").
	SetValue("something-secure").
	SetARN("arn:aws:ssm:us-east-2:aws-account-id:/my-service/dev/DB_PASSWORD")

var param2 = new(ssm.Parameter).
	SetName("/my-service/dev/DB_HOST").
	SetValue("rds.something.aws.com").
	SetARN("arn:aws:ssm:us-east-2:aws-account-id:/my-service/dev/DB_HOST")

var param3 = new(ssm.Parameter).
	SetName("/my-service/dev/DB_USERNAME").
	SetValue("username").
	SetARN("arn:aws:ssm:us-east-2:aws-account-id:/my-service/dev/DB_USERNAME")

var errSSM = errors.New("ssm request error")

type stubGetParametersByPathOutput struct {
	Output         ssm.GetParametersByPathOutput
	MoreParamsLeft bool
}

type stubSSMClient struct {
	GetParametersByPathOutput        []stubGetParametersByPathOutput
	GetParametersByPathError         error
	GetParameterOutput               *ssm.GetParameterOutput
	GetParameterError                error
}

func (s stubSSMClient) GetParametersByPathPages(input *ssm.GetParametersByPathInput, fn func(*ssm.GetParametersByPathOutput, bool) bool) error {
	if s.GetParametersByPathError == nil {
		for _, output := range s.GetParametersByPathOutput {
			done := fn(&output.Output, output.MoreParamsLeft)
			if done {
				return nil
			}
		}
	}
	return s.GetParametersByPathError
}

func (s stubSSMClient) GetParameter(input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	return s.GetParameterOutput, s.GetParameterError
}

func TestClient_GetParametersByPath(t *testing.T) {
	tests := []struct {
		name           string
		ssmClient      ssmClient
		path           string
		expectedError  error
		expectedOutput *Parameters
	}{
		{
			name: "Success",
			ssmClient: &stubSSMClient{
				GetParametersByPathOutput: []stubGetParametersByPathOutput{
					{
						MoreParamsLeft: true,
						Output: ssm.GetParametersByPathOutput{
							Parameters: getParameters(),
						},
					},
					{
						MoreParamsLeft: false,
						Output: ssm.GetParametersByPathOutput{
							Parameters: []*ssm.Parameter{
								param3,
							},
						},
					},
				},
			},
			path: "/my-service/dev/",
			expectedOutput: &Parameters{
				basePath: "/my-service/dev/",
				parameters: map[string]*Parameter{
					"/my-service/dev/DB_PASSWORD": {Value: param1.Value},
					"/my-service/dev/DB_HOST":     {Value: param2.Value},
					"/my-service/dev/DB_USERNAME": {Value: param3.Value},
				},
			},
		},

		{
			name: "Failed SSM Request Error",
			ssmClient: &stubSSMClient{
				GetParametersByPathError: errSSM,
			},
			path:          "/my-service/dev/",
			expectedError: errSSM,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			client := NewParameterStoreWithClient(test.ssmClient)
			parameters, err := client.GetAllParametersByPath(test.path, true)
			if err != test.expectedError {
				t.Errorf(`Unexpected error: got %d, expected %d`, err, test.expectedError)
			}
			if !reflect.DeepEqual(parameters, test.expectedOutput) {
				t.Error(`Unexpected parameters`, *parameters, *test.expectedOutput)
			}
		})
	}
}

func getParameters() []*ssm.Parameter {
	return []*ssm.Parameter{
		param1, param2,
	}
}

func TestParameterStore_GetParameter(t *testing.T) {
	value := "something-secure"
	tests := []struct {
		name           string
		ssmClient      ssmClient
		parameterName  string
		expectedError  error
		expectedOutput *Parameter
	}{
		{
			name: "Success",
			ssmClient: &stubSSMClient{
				GetParameterOutput: &ssm.GetParameterOutput{
					Parameter: param1,
				},
			},
			parameterName: "/my-service/dev/DB_PASSWORD",
			expectedOutput: &Parameter{
				Value: &value,
			},
		},
		{
			name:          "Failed Empty name",
			ssmClient:     &stubSSMClient{},
			parameterName: "",
			expectedError: ErrParameterInvalidName,
		},
		{
			name: "Failed Parameter Not Found",
			ssmClient: &stubSSMClient{
				GetParameterError: awserr.New(ssm.ErrCodeParameterNotFound, "parameter not found", nil),
			},
			parameterName: "/my-service/dev/NOT_FOUND",
			expectedError: ErrParameterNotFound,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			client := NewParameterStoreWithClient(test.ssmClient)
			parameter, err := client.GetParameter(test.parameterName, true)
			if err != test.expectedError {
				t.Errorf(`Unexpected error: got %d, expected %d`, err, test.expectedError)
			}
			if !reflect.DeepEqual(parameter, test.expectedOutput) {
				t.Error(`Unexpected parameter`, *parameter, *test.expectedOutput)
			}
		})
	}
}
