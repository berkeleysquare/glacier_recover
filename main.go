package main

import (
    "crypto/tls"
    "fmt"
    "github.com/SpectraLogic/glacier_recover/commands"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/awserr"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/glacier"
    "github.com/aws/aws-sdk-go/service/s3"
    "net/http"
)


func printAwsErr (err error) {
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            switch aerr.Code() {
            case glacier.ErrCodeResourceNotFoundException:
                fmt.Println(glacier.ErrCodeResourceNotFoundException, aerr.Error())
            case glacier.ErrCodeInvalidParameterValueException:
                fmt.Println(glacier.ErrCodeInvalidParameterValueException, aerr.Error())
            case glacier.ErrCodeMissingParameterValueException:
                fmt.Println(glacier.ErrCodeMissingParameterValueException, aerr.Error())
            case glacier.ErrCodeServiceUnavailableException:
                fmt.Println(glacier.ErrCodeServiceUnavailableException, aerr.Error())
            default:
                fmt.Println(aerr.Error())
            }
        } else {
            // Print the error, cast err to awserr.Error to get the Code and
            // Message from an error.
            fmt.Println(err.Error())
        }
        return
    }
}

// const region = "us-west-2"
// const vaultName = "jk-neo-vault"
// const vaultName = "AWS-Vail-bucket"
// const accountId = "-"
// const jobId = "8dtTBQk_G30bHKlE9JbvB1ctJtM8wemo_8vY3z06QcMDPcL3Q7C8Ri_1NvhueX8AprXkfIRUosI7D_nhPa8GP-SjbLce"
// const jobId = "8ENxgOWQEZGA0Xnk55fPi1aHyf0Fzj1KXSwpli8fN7xXc3JyILNtxsqDZep79YPgbm9RuB8PkSL__sfcXIEp2eS8MQll"

func main() {

    // Parse the arguments.
    args, argsErr := commands.ParseArgs()
    if argsErr != nil {
        commands.ListCommands(args)
        printAwsErr(argsErr)
        return
    }

    if args.Command == "list_commands" || args.Command == "" {
        commands.ListCommands(args)
        return
    }

    // Create a Glacier client from just a session.
    tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: args.NoVerifySSL},}
    httpClient := &http.Client{Transport: tr}
    mySession := session.Must(session.NewSessionWithOptions( session.Options{Profile: args.Profile }))
    svc := s3.New(mySession,
        aws.NewConfig().WithHTTPClient(httpClient).WithS3ForcePathStyle(true).WithRegion(args.Region).WithEndpoint(args.Endpoint))

    // Run the command
    err := commands.RunCommand(svc, args)
    if err != nil {
        printAwsErr(err)
        return
    }

    fmt.Printf("Ready\n",)
}



