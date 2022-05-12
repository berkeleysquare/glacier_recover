package commands

import (
    "fmt"
    "github.com/aws/aws-sdk-go/service/s3"
)

type command func(*s3.S3, *Arguments) error

var availableCommands = map[string]command {
    "list_buckets": getBucketList,
    "inventory": getBucketInventory,
    "restore": restoreObject,
    "get_object": getObject,
    "delete_object": deleteObject,
    "get_object_byte": getObjectByte,
    "test_byte_restore": testByteRestore,
    "head_object": headObject,
    "restore_from_glacier": restoreFromGlacier,
}

func RunCommand(svc *s3.S3, args *Arguments) error {
    cmd, ok := availableCommands[args.Command]
    if ok {
        return cmd(svc, args)
    } else {
        return fmt.Errorf("Unsupported command: '%s'", args.Command)
    }
}

func ListCommands(args *Arguments) error {
    fmt.Printf("Usage: recover_glacier --command <command>\n",)
    for key, _ := range availableCommands {
        fmt.Printf("%s\n", key)
    }
    return nil
}
