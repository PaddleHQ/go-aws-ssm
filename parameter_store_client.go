package awsssm

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type ssmClient interface {
	GetParametersByPath(input *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error)
}

//ParameterStore holds all the methods tha are supported against AWS Parameter Store
type ParameterStore struct {
	ssm ssmClient
}

//SetSSM sets a new ssm client
func (c *ParameterStore) SetSSM(ssm ssmClient) {
	c.ssm = ssm
}

//GetAllParametersByPath is returning all the Parameters that are hierarchy linked to this path
//For example a request with path as /my-service/dev/
//Will return /my-service/dev/param-a, /my-service/dev/param-b, etc... but will not return recursive paths
//the `ssm:GetAllParametersByPath` permission is required
//to the `arn:aws:ssm:us-east-2:aws-account-id:/my-service/dev/*`
func (c *ParameterStore) GetAllParametersByPath(path string, decrypt bool) (*Parameters, error) {
	var input = &ssm.GetParametersByPathInput{}
	input.SetWithDecryption(decrypt)
	input.SetPath(path)
	return c.getParameters(input)
}

func (c *ParameterStore) getParameters(input *ssm.GetParametersByPathInput) (*Parameters, error) {
	result, err := c.ssm.GetParametersByPath(input)
	if err != nil {
		return nil, err
	}
	parameters := NewParameters(*input.Path, make(map[string]*Parameter, len(result.Parameters)))
	for _, v := range result.Parameters {
		if v.Name == nil {
			continue
		}
		parameters.parameters[*v.Name] = &Parameter{ssmParameter: v}
	}
	return parameters, nil
}

var defaultSessionOptions = session.Options{
	Config:            aws.Config{Region: aws.String("us-east-1")},
	SharedConfigState: session.SharedConfigEnable,
}

//NewParameterStoreWithOptions is creating a new parameterstore ParameterStore with the specified Session options
func NewParameterStoreWithOptions(sessionOptions *session.Options, ssmConfig ...*aws.Config) (*ParameterStore, error) {
	options := defaultSessionOptions
	if sessionOptions != nil {
		options = *sessionOptions
	}
	sessionAWS, err := session.NewSessionWithOptions(options)
	if err != nil {
		return nil, err
	}
	svc := ssm.New(sessionAWS, ssmConfig...)
	return &ParameterStore{ssm: svc}, nil
}

//NewParameterStore is creating a new parameterstore ParameterStore
func NewParameterStore(ssmConfig ...*aws.Config) (*ParameterStore, error) {
	sessionAWS, err := session.NewSessionWithOptions(defaultSessionOptions)
	if err != nil {
		return nil, err
	}
	svc := ssm.New(sessionAWS, ssmConfig...)
	return &ParameterStore{ssm: svc}, nil
}
