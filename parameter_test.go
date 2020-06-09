package awsssm

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"
	"time"
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

func TestParameters_Read(t *testing.T) {
	tests := []struct {
		name              string
		basePath          string
		parameters        map[string]*Parameter
		expectedError     error
		expectedByteCount int64
	}{
		{
			name:              "ReadFrom Small Byte Count",
			basePath:          "/my-service/dev",
			parameters:        getParametersMap(),
			expectedError:     nil,
			expectedByteCount: 70,
		},
		{
			name:              "ReadFrom Large Byte Count",
			basePath:          "/my-service/dev",
			parameters:        getRandomParametersMap(1000, 10),
			expectedError:     nil,
			expectedByteCount: 33001,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parameter := NewParameters(test.basePath, test.parameters)
			buf := new(bytes.Buffer)
			bytesRead, err := buf.ReadFrom(parameter)
			if err != test.expectedError {
				t.Errorf(`Unexpected error: got %s, expected %s`, err, test.expectedError)
			}
			if bytesRead != test.expectedByteCount {
				t.Errorf(`Unexpected value: got %d, expected %d`, bytesRead, test.expectedByteCount)
			}
		})
	}
}

func getParametersMap() map[string]*Parameter {
	return map[string]*Parameter{
		"/my-service/dev/DB_PASSWORD": {Value: param1.Value},
		"/my-service/dev/DB_HOST":     {Value: param2.Value},
	}
}

var seededRand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func getRandomParametersMap(n int, l int) map[string]*Parameter {
	m := make(map[string]*Parameter)
	for i := 0; i < n; i++ {
		s := "/my-service/dev/"
		for j := 0; j < l; j++ {
			char := charset[seededRand.Intn(int(^uint(0)>>1))%len(charset)]
			s += string(char)
		}
		m[s] = &Parameter{
			Value: param1.Value,
		}
	}
	return m
}
