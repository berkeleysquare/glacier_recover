package commands

import (
    "flag"
    "os"
)

// Represents the parsed command line arguments that we may be interested in.
type Arguments struct {
    Endpoint, Proxy string
    Region string
    Profile string
    Bucket string
    Command string
    Key string
    Prefix string
    Download bool
    NoVerifySSL bool
    DeleteOnFail bool
    OutputFile string
}

func ParseArgs() (*Arguments, error) {
    // Parse command line arguments.
    endpointParam := flag.String("endpoint", "", "Specifies the url to the DS3 server.")
    proxyParam := flag.String("proxy", "", "Specifies the HTTP proxy to route through.")
    regionParam := flag.String("region", "us-west-2", "Specifies the S3 region.")
    profileParam := flag.String("profile", "default", "AWS CLI profile.")
    commandParam := flag.String("command", "", "The call to execute: use list_commands for valid commands")
    bucketParam := flag.String("bucket", "", "The name of the bucket to constrict the request to.")
    keyParam := flag.String("key", "", "Object name (key).")
    prefixParam := flag.String("prefix", "", "Match objects starting with prefix.")
    downloadParam := flag.Bool("download", true, "True to download after recovery")
    noVerifySslParam := flag.Bool("no-verify-ssl", false, "True to allow self-signed certificates")
    deleteOnFailParam := flag.Bool("delete-on-fail", false, "True to delete on get_object_byte fails")
    outputFile:= flag.String("out", "", "output file path")
    flag.Parse()

    // Build the arguments object.
    args := Arguments{
        Endpoint: paramOrEnv(*endpointParam, "DS3_ENDPOINT"),
        Proxy: paramOrEnv(*proxyParam, "DS3_PROXY"),
        Region: paramOrEnv(*regionParam, "AWS_REGION"),
        Profile: *profileParam,
        Bucket: *bucketParam,
        Command: *commandParam,
        Key: *keyParam,
        Prefix: *prefixParam,
        Download: *downloadParam,
        NoVerifySSL: *noVerifySslParam,
        DeleteOnFail: *deleteOnFailParam,
        OutputFile: *outputFile,
    }
    return &args, nil
}

func paramOrEnv(param, envName string) string {
    env := os.Getenv(envName)
    switch {
    case param != "": return param
    case env != "": return env
    default: return ""
    }
}
