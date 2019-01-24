package awsssm

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/mitchellh/mapstructure"
)

//Parameter holds a Systems Manager parameter from AWS Parameter Store
type Parameter struct {
	ssmParameter *ssm.Parameter
}

//GetValue return the actual value of the parameter
func (p *Parameter) GetValue() string {
	if p.ssmParameter.Value == nil {
		return ""
	}
	return *p.ssmParameter.Value
}

//NewParameters creates a Parameters
func NewParameters(basePath string, parameters map[string]*Parameter) *Parameters {
	return &Parameters{
		basePath:   basePath,
		parameters: parameters,
	}
}

//Parameters holds the output and all AWS Parameter Store that have the same base path
type Parameters struct {
	basePath   string
	parameters map[string]*Parameter
}

//GetValueByName returns the value based on the name
//so the AWS Parameter Store parameter name is base path + name
func (p *Parameters) GetValueByName(name string) string {
	parameter, ok := p.parameters[p.basePath+name]
	if !ok {
		return ""
	}
	return parameter.GetValue()
}

//GetValueByFullPath returns the value based on the full path
func (p *Parameters) GetValueByFullPath(name string) string {
	parameter, ok := p.parameters[name]
	if !ok {
		return ""
	}
	return parameter.GetValue()
}

//Decode decodes the parameters into the given struct
//We are using this package to decode the values to the struct https://github.com/mitchellh/mapstructure
//For more details how you can use this check the parameter_test.go file
func (p *Parameters) Decode(output interface{}) error {
	return mapstructure.Decode(p.getKeyValueMap(), output)
}

//JSONMarshal marshals in json format the key value map
//This can be used with bytes.NewBuffer to create a new io.Reader
//and then used with tools like viper https://github.com/spf13/viper
func (p *Parameters) JSONMarshal() ([]byte, error) {
	bytes.NewBuffer([]byte{})
	return json.Marshal(p.getKeyValueMap())
}

func (p *Parameters) getKeyValueMap() map[string]string {
	keyValue := make(map[string]string, len(p.parameters))
	for k, v := range p.parameters {
		keyValue[strings.Replace(k, p.basePath, "", 1)] = v.GetValue()
	}
	return keyValue
}

//JSONMarshallParameters will return the error if not nil or the json format of the key/value
//for example:
// 	paramsByte, err := awsssm.JSONMarshallParameters(
//		pmstore.GetAllParametersByPath("/my-service/dev/", true),
//	)
//	if err != nil {
//		return err
//	}
func JSONMarshallParameters(parameters *Parameters, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	return parameters.JSONMarshal()
}
