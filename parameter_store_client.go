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
	PutParameter(input *ssm.PutParameterInput) (*ssm.PutParameterOutput, error)
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
//
//This will also page through and return all elements in the hierarchy, non-recursively
func (ps *ParameterStore) GetAllParametersByPath(path string, decrypt bool) (*Parameters, error) {
	var input = &ssm.GetParametersByPathInput{}
	input.SetWithDecryption(decrypt)
	input.SetPath(path)
  input.SetMaxResults(10)
	return ps.getParameters(input)
}

func (ps *ParameterStore) getParameters(input *ssm.GetParametersByPathInput) (*Parameters, error) {
  allParams := make([]*ssm.Parameter, 0)
  for {
    result, err := ps.ssm.GetParametersByPath(input)
    if err != nil {
      return nil, err
    }
    allParams = append(allParams, result.Parameters...)

    if result.NextToken != nil {
      input.SetNextToken(*result.NextToken)
    } else {
      break
    }
  }
	parameters := NewParameters(*input.Path, make(map[string]*Parameter, len(allParams)))
	for _, v := range allParams {
		if v.Name == nil {
			continue
		}
		parameters.parameters[*v.Name] = &Parameter{Value: v.Value}
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

func (ps *ParameterStore) PutSecureParameter(name, value string, overwrite bool) error {
  return ps.putSecureParameterWrapper(name, value, "", overwrite)
}
func (ps *ParameterStore) PutSecureParameterWithCMK(name, value string, overwrite bool, kmsID string) error {
  return ps.putSecureParameterWrapper(name, value, kmsID, overwrite)
}
func (ps *ParameterStore) putSecureParameterWrapper(name, value, kmsID string, overwrite bool) error {
	if name == "" {
		return ErrParameterInvalidName
	}
	input := &ssm.PutParameterInput{}
	input.SetName(name)
	input.SetType("SecureString")
  input.SetValue(value)
	if kmsID != "" {
		input.SetKeyId(kmsID)
	}
  input.SetOverwrite(overwrite)

  if err := input.Validate(); err != nil {
    return err
  }

	return ps.putParameter(input)
}
func (ps *ParameterStore) putParameter(input *ssm.PutParameterInput) error {
	_, err := ps.ssm.PutParameter(input)
	if err != nil {
    if awsError, ok := err.(awserr.Error); ok && awsError.Code() == ssm.ErrCodeParameterNotFound {
			return ErrParameterNotFound
		}
		return err
	}
	return nil
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
