package awsssm

import (
	"errors"
	"fmt"
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

//  return s.GetParametersByPathOutput, s.GetParametersByPathError
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
	GetParametersByPathOutput []stubGetParametersByPathOutput
	GetParametersByPathError  error
	GetParameterOutput        *ssm.GetParameterOutput
	GetParameterError         error
	PutParameterInputReceived ssm.PutParameterInput
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

// we return nothing becuase the actual response is pretty boring. Just a version number. We DO
// want to track was is input because there is a _little_ business logic around that
func (s stubSSMClient) PutParameter(input *ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
	s.PutParameterInputReceived = *input
	fmt.Printf("PutParameterInputReceived: %v", *input)
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
				GetParametersByPathOutput: []stubGetParametersByPathOutput{
					{
						MoreParamsLeft: true,
						Output: ssm.GetParametersByPathOutput{
							Parameters: getParameters(),
						},
					},
					{
						MoreParamsLeft: true,
						Output: ssm.GetParametersByPathOutput{
							Parameters: getParameters2(),
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
				t.Errorf(`Unexpected parameters: got: %+v, expected: %+v`, *parameters, *test.expectedOutput)
			}
		})
	}
}

func getParameters() []*ssm.Parameter {
	return []*ssm.Parameter{
		param1, param2,
	}
}

func getParameters2() []*ssm.Parameter {
	return []*ssm.Parameter{
		param3,
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
	paramType := "SecureString"
	overwriteTrue := true
	overwriteFalse := false
	tests := []struct {
		name           string
		ssmClient      ssmClient
		parameterName  string
		parameterValue string
		overwrite      bool
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
				Name:      &paramName,
				Type:      &paramType,
				Value:     &paramValue,
				Overwrite: &overwriteFalse,
			},
		},
		{
			name:           "Overwrite Changes Propagate",
			ssmClient:      &stubSSMClient{},
			parameterName:  paramName,
			parameterValue: paramValue,
			overwrite:      overwriteTrue,
			expectedOutput: ssm.PutParameterInput{
				Name:      &paramName,
				Type:      &paramType,
				Value:     &paramValue,
				Overwrite: &overwriteTrue,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := NewParameterStoreWithClient(test.ssmClient)
			err := client.PutSecureParameter(test.parameterName, test.parameterValue, test.overwrite)
			if err != test.expectedError {
				t.Errorf(`Unexpected error: got %d, expected %d`, err, test.expectedError)
			}
			fullStubClient := test.ssmClient.(*stubSSMClient)
			fmt.Printf("fullStubClient: %+v", fullStubClient)
			if !reflect.DeepEqual((*fullStubClient).PutParameterInputReceived, test.expectedOutput) {
				t.Errorf(`Unexpected parameter: got %v, expected %v`, (*fullStubClient).PutParameterInputReceived, test.expectedOutput)
			}
		})
	}
}

func TestParameterStore_PutSecureParameterWithCMK(t *testing.T) {
	paramName := "foo"
	paramValue := "baz"
	paramType := "SecureString"
	overwriteFalse := false
	kmsID := "super-secret-kms"
	tests := []struct {
		name           string
		ssmClient      ssmClient
		parameterName  string
		parameterValue string
		overwrite      bool
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
				Name:      &paramName,
				Overwrite: &overwriteFalse,
				Type:      &paramType,
				Value:     &paramValue,
			},
		},
		{
			name:           "KMS ID Changes Propagate",
			ssmClient:      &stubSSMClient{},
			parameterName:  paramName,
			parameterValue: paramValue,
			kmsID:          kmsID,
			expectedOutput: ssm.PutParameterInput{
				KeyId:     &kmsID,
				Name:      &paramName,
				Overwrite: &overwriteFalse,
				Type:      &paramType,
				Value:     &paramValue,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := NewParameterStoreWithClient(test.ssmClient)
			err := client.PutSecureParameterWithCMK(test.parameterName, test.parameterValue, test.overwrite, test.kmsID)
			if err != test.expectedError {
				t.Errorf(`Unexpected error: got %d, expected %d`, err, test.expectedError)
			}
			fullStubClient := test.ssmClient.(*stubSSMClient)
			if !reflect.DeepEqual(fullStubClient.PutParameterInputReceived, test.expectedOutput) {
				t.Error(`Unexpected parameter`, fullStubClient.PutParameterInputReceived, test.expectedOutput)
			}
		})
	}
}
