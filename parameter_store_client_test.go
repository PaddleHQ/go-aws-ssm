package awsssm

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
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

var putParameterInputReceived ssm.PutParameterInput

type stubSSMClient struct {
	GetParametersByPathOutput *ssm.GetParametersByPathOutput
	GetParametersByPathError  error
	GetParameterOutput        *ssm.GetParameterOutput
	GetParameterError         error
}

func (s stubSSMClient) GetParametersByPath(input *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error) {
	return s.GetParametersByPathOutput, s.GetParametersByPathError
}

func (s stubSSMClient) GetParameter(input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	return s.GetParameterOutput, s.GetParameterError
}

// we don't really use this because there isn't much to actually test for PutParameter
// it accepts an input and either returns an error or nil--that's it
func (s stubSSMClient) PutParameter(input *ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
	putParameterInputReceived = *input
	return nil, nil
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

func TestParameterStore_PutSecureParameter(t *testing.T) {
	paramName := "foo"
	paramValue := "baz"
	kmsId := "bar"
	paramType := "SecureString"
	tests := []struct {
		name           string
		ssmClient      ssmClient
		parameterName  string
		parameterValue string
		kmsID          string
		expectedError  error
		expectedOutput ssm.PutParameterInput
	}{
		{
			name:           "Failed Empty name",
			ssmClient:      &stubSSMClient{},
			parameterName:  "",
			parameterValue: "",
			expectedError:  ErrParameterInvalidName,
		},
		{
			name:           "Set Correct Defaults",
			ssmClient:      &stubSSMClient{},
			parameterName:  paramName,
			parameterValue: paramValue,
			expectedOutput: ssm.PutParameterInput{
				Name:  &paramName,
				KeyId: nil,
				Type:  &paramType,
				Value: &paramValue,
			},
		},
		{
			name:           "Set Correct KMS ID",
			ssmClient:      &stubSSMClient{},
			parameterName:  paramName,
			parameterValue: paramValue,
			kmsID:          kmsId,
			expectedOutput: ssm.PutParameterInput{
				Name:  &paramName,
				KeyId: &kmsId,
				Type:  &paramType,
				Value: &paramValue,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			client := NewParameterStoreWithClient(test.ssmClient)
			err := client.PutSecureParameter(test.parameterName, test.parameterValue, test.kmsID)
			if err != test.expectedError {
				t.Errorf(`Unexpected error: got %d, expected %d`, err, test.expectedError)
			}
			if !reflect.DeepEqual(putParameterInputReceived, test.expectedOutput) {
				t.Error(`Unexpected parameter`, putParameterInputReceived, test.expectedOutput)
			}
		})
	}
}

//func TestParameterStore_PutSecureParameter_SetsSecureStringType(t *testing.T) {
//}

//func TestParameterStore_PutSecureParameter_PassesKMSIDIfNotEmpty(t *testing.T) {
//}

//func TestParameterStore_PutSecureParameter_DoesNotPassKMSIDIfEmpty(t *testing.T) {
//}
