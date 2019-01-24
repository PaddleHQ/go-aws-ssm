# go-aws-ssm
Client library that interfaces with AWS System Manager Agent

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
    	params, err := pmstore.GetAllParametersByPath("/my-service/dev/")
    	if err!=nil{
    		return err
    	}
    	
    	//And getting a specific value
    	value:=params.GetValueByName("param-1")
    	//value should be `a`
    	
    	
```

#### Integrates easy with [viper](https://github.com/spf13/viper)
```go
        //Assuming you have the parameters in the following format:
    	//my-service/dev/param-1  -> with value `a`
    	//my-service/dev/param-2  -> with value `b`
    	pmstore, err := awsssm.NewParameterStore()
    	if err != nil {
    		return err
    	}
    	//Requesting the base path and json marshaling the key/value of parameters
    	paramsByte, err := ssm.JSONMarshallParameters(
    		pmstore.GetAllParametersByPath("/my-service/dev/", true),
    	)
    	if err != nil {
    		return err
    	}
    
    	//Configure viper to handle it as json document, nothing special here!
    	v := viper.New()
    	v.SetConfigType(`json`)
    	err = v.ReadConfig(bytes.NewBuffer(paramsByte))
    	if err != nil {
    		return err
    	}
    	value := v.Get(`param-1`)
    	//value should be `a`
```
