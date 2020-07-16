package awsssm

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

var (
	//ErrParameterNotFound error for when the requested Parameter Store parameter can't be found
	ErrParameterNotFound = errors.New("parameter not found")
	//ErrParameterInvalidName error for invalid parameter name
	ErrParameterInvalidName = errors.New("invalid parameter name")
)

type ssmClient interface {
	GetParametersByPath(input *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error)
	GetParameter(input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
}

//ParameterStore holds all the methods tha are supported against AWS Parameter Store
type ParameterStore struct {
	ssm ssmClient
}

//GetAllParametersByPath is returning all the Parameters that are hierarchy linked to this path
//For example a request with path as /my-service/dev/
//Will return /my-service/dev/param-a, /my-service/dev/param-b, etc... but will not return recursive paths
//the `ssm:GetAllParametersByPath` permission is required
//to the `arn:aws:ssm:aws-region:aws-account-id:/my-service/dev/*`
func (ps *ParameterStore) GetAllParametersByPath(path string, decrypt bool) (*Parameters, error) {
	var input = &ssm.GetParametersByPathInput{}
	input.SetWithDecryption(decrypt)
	input.SetPath(path)
	return ps.getParameters(input)
}

func (ps *ParameterStore) getParameters(input *ssm.GetParametersByPathInput) (*Parameters, error) {
	result, err := ps.ssm.GetParametersByPath(input)
	hasNext := false
	if err != nil {
		return nil, err
	}
	parameters := NewParameters(*input.Path, make(map[string]*Parameter, len(result.Parameters)))
	for hasNext {
		for _, v := range result.Parameters {
			if v.Name == nil {
				continue
			}
			parameters.parameters[*v.Name] = &Parameter{Value: v.Value}
		}
		if result.NextToken != nil{
			input.NextToken = result.NextToken
			result, err = ps.ssm.GetParametersByPath(input)
			if err != nil {
				return nil, err
			}
		}else{
			hasNext = false
		}
	}
	return parameters, nil
}

//GetParameter is returning the parameter with the given name
//For example a request with name as /my-service/dev/param-1
//Will return the parameter value if exists or ErrParameterInvalidName if parameter cannot be found
//The `ssm:GetParameter` permission is required
//to the `arn:aws:ssm:aws-region:aws-account-id:/my-service/dev/param-1` resource
func (ps *ParameterStore) GetParameter(name string, decrypted bool) (*Parameter, error) {
	if name == "" {
		return nil, ErrParameterInvalidName
	}
	input := &ssm.GetParameterInput{}
	input.SetName(name)
	input.SetWithDecryption(decrypted)
	return ps.getParameter(input)
}
func (ps *ParameterStore) getParameter(input *ssm.GetParameterInput) (*Parameter, error) {
	result, err := ps.ssm.GetParameter(input)
	if err != nil {
		if awsError, ok := err.(awserr.Error); ok && awsError.Code() == ssm.ErrCodeParameterNotFound {
			return nil, ErrParameterNotFound
		}
		return nil, err
	}
	return &Parameter{
		Value: result.Parameter.Value,
	}, nil
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
