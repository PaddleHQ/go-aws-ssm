package awsssm

import (
	"errors"
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

var errSSM = errors.New("ssm request error")

type stubSSMClient struct {
	GetParametersByPathOutput *ssm.GetParametersByPathOutput
	GetParametersByPathError  error
}

func (s stubSSMClient) GetParametersByPath(input *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error) {
	return s.GetParametersByPathOutput, s.GetParametersByPathError
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
				GetParametersByPathOutput: &ssm.GetParametersByPathOutput{
					Parameters: getParameters(),
				},
			},
			path: "/my-service/dev/",
			expectedOutput: &Parameters{
				basePath: "/my-service/dev/",
				parameters: map[string]*Parameter{
					"/my-service/dev/DB_PASSWORD": {Value: param1.Value},
					"/my-service/dev/DB_HOST":     {Value: param2.Value},
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

			client := new(ParameterStore)
			client.SetSSM(test.ssmClient)
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
