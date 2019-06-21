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
		parameters.parameters[*v.Name] = &Parameter{Value: v.Value}
	}
	return parameters, nil
}

//NewParameterStoreWithClient is creating a new ParameterStore with the given ssm Client
func NewParameterStoreWithClient(client ssmClient) *ParameterStore {
	return &ParameterStore{ssm: client}
}

//NewParameterStore is creating a new ParameterStore by creating an AWS Session
func NewParameterStore(ssmConfig ...*aws.Config) (*ParameterStore, error) {
	sessionAWS, err := session.NewSession(ssmConfig...)
	if err != nil {
		return nil, err
	}
	svc := ssm.New(sessionAWS)
	return &ParameterStore{ssm: svc}, nil
}
