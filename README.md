[![Build Status](https://travis-ci.com/PaddleHQ/go-aws-ssm.svg?branch=master)](https://travis-ci.com/PaddleHQ/go-aws-ssm)
[![codecov](https://codecov.io/gh/PaddleHQ/go-aws-ssm/branch/master/graph/badge.svg)](https://codecov.io/gh/PaddleHQ/go-aws-ssm)
[![Go Report Card](https://goreportcard.com/badge/github.com/PaddleHQ/go-aws-ssm)](https://goreportcard.com/report/github.com/PaddleHQ/go-aws-ssm)
[![GoDoc](https://godoc.org/github.com/PaddleHQ/go-aws-ssm?status.svg)](https://pkg.go.dev/github.com/PaddleHQ/go-aws-ssm)

# go-aws-ssm
Go package that interfaces with [AWS System Manager](https://www.amazonaws.cn/en/systems-manager/).

## Why to use go-aws-ssm and not the aws-sdk-go?
This package is wrapping the aws-sdk-go and hides the complexity dealing with the not so Go friendly AWS SDK.
Perfect use case for this package is when secure parameters for an application are stored to 
[AWS Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html)
using a path hierarchy. During application startup you can use this package to fetch them and use them in your application.

## Install

```bash
go get github.com/PaddleHQ/go-aws-ssm
```

## Examples 

#### Basic Usage

```go
        //Assuming you have the parameters in the following format:
    	//my-service/dev/param-1  -> with value `a`
    	//my-service/dev/param-2  -> with value `b`
    	pmstore, err := awsssm.NewParameterStore()
    	if err != nil {
    		return err
    	}
    	//Requesting the base path
    	params, err := pmstore.GetAllParametersByPath("/my-service/dev/", true)
    	if err!=nil{
    		return err
    	}
    	
    	//And getting a specific value
    	value:=params.GetValueByName("param-1")
    	//value should be `a`
    	
    	
```

#### Integrates easily with [viper](https://github.com/spf13/viper)
```go
        //Assuming you have the parameters in the following format:
     	//my-service/dev/param-1  -> with value `a`
     	//my-service/dev/param-2  -> with value `b`
     	pmstore, err := awsssm.NewParameterStore()
     	if err != nil {
     		return err
     	}
     	//Requesting the base path
     	params, err := pmstore.GetAllParametersByPath("/my-service/dev/", true)
     	if err!=nil{
     		return err
     	}
    
    	//Configure viper to handle it as json document, nothing special here!
    	v := viper.New()
    	v.SetConfigType(`json`)
    	//params object implements the io.Reader interface that is required
    	err = v.ReadConfig(params)
    	if err != nil {
    		return err
    	}
    	value := v.Get(`param-1`)
    	//value should be `a`
```
