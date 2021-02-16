package awsssm

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/mitchellh/mapstructure"
)

//Parameter holds a Systems Manager parameter from AWS Parameter Store
type Parameter struct {
	Value *string
}

//GetValue return the actual Value of the parameter
func (p *Parameter) GetValue() string {
	if p.Value == nil {
		return ""
	}
	return *p.Value
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
	readIndex  int64
	bytesJSON  []byte
	basePath   string
	parameters map[string]*Parameter
}

//Read implements the io.Reader interface for the key/value pair
func (p *Parameters) Read(des []byte) (n int, err error) {
	if p.bytesJSON == nil {
		p.bytesJSON, err = json.Marshal(p.getKeyValueMap())
		if err != nil {
			return 0, err
		}
	}

	if p.readIndex >= int64(len(p.bytesJSON)) {
		p.readIndex = 0
		return 0, io.EOF
	}

	n = copy(des, p.bytesJSON[p.readIndex:])
	p.readIndex += int64(n)

	return n, nil
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

func (p *Parameters) getKeyValueMap() map[string]string {
	keyValue := make(map[string]string, len(p.parameters))
	for k, v := range p.parameters {
		keyValue[strings.Replace(k, p.basePath, "", 1)] = v.GetValue()
	}
	return keyValue
}

// GetAllValues returns a map with all the keys and values in the store.
func (p *Parameters) GetAllValues() map[string]string {
	return p.getKeyValueMap()
}
