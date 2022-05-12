package commands

import (
    "bufio"
    "bytes"
    "encoding/csv"
    "fmt"
    "github.com/SpectraLogic/glacier_recover/client"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/s3"
    "io"
    "log"
    "os"
    "path"
    "sync"
    "time"
)


func getBucketList(svc *s3.S3, args *Arguments) error {
    outputFile := args.OutputFile
    wOut := os.Stdout
    if len(outputFile) > 0 {
        f, err := os.Create(outputFile)
        if err != nil {
            return fmt.Errorf("Could not create %s\n%v\n", outputFile, err)
        }
        defer f.Close()
        wOut = f
    }
    w := csv.NewWriter(wOut)
    defer w.Flush()

    vail := &client.VailClient{ svc, w, "", "", false}

    vail.PrintListBucketsCsvHeader()
    bucketList, err := svc.ListBuckets(nil)
    if err == nil {
        return vail.PrintBucketList(bucketList.Buckets)
    }
    return err
}

func getBucketInventory(svc *s3.S3, args *Arguments) error {
    return paginatedBucketInventory(svc, args.Bucket, args.Prefix, args.OutputFile)
}

func doBucketInventory(svc *s3.S3, bucket string, prefix string) ([]*s3.Object, error) {
    bucketObjectListResponse, err := svc.ListObjectsV2(
        &s3.ListObjectsV2Input{
            Bucket: aws.String(bucket),
            Prefix: aws.String(prefix)})
    if err == nil {
        return bucketObjectListResponse.Contents, nil
    }
    return nil, err
}


func paginatedBucketInventory(svc *s3.S3, bucket string, prefix string, outputFile string) error {
    wOut := os.Stdout
    if len(outputFile) > 0 {
        f, err := os.Create(outputFile)
        if err != nil {
            return fmt.Errorf("Could not create %s\n%v\n", outputFile, err)
        }
        defer f.Close()
        wOut = f
    }
    w := csv.NewWriter(wOut)
    defer w.Flush()

    vail := &client.VailClient{ svc, w, bucket, prefix, false}

    err := vail.PrintBucketObjectsCsvHeader()
    if err != nil {
        return fmt.Errorf("failed printing header %v\n", err)
    }
    return svc.ListObjectsV2Pages(
        &s3.ListObjectsV2Input{
        Bucket: aws.String(bucket),
        Prefix: aws.String(prefix),
        MaxKeys: aws.Int64(100)},
        vail.PrintObjectsPage)
}

func testByteRestore(svc *s3.S3, args *Arguments) error {
    return doTestByteRestore(svc, args.Bucket, args.Prefix, args.DeleteOnFail, args.OutputFile)
}

func doTestByteRestore(svc *s3.S3, bucket string, prefix string, deleteOnFail bool, outputFile string) error {
    wOut := os.Stdout
    if len(outputFile) > 0 {
        f, err := os.Create(outputFile)
        if err != nil {
            return fmt.Errorf("Could not create %s\n%v\n", outputFile, err)
        }
        defer f.Close()
        wOut = f
    }
    w := csv.NewWriter(wOut)
    defer w.Flush()

    vail := &client.VailClient{ svc, w, bucket, prefix, deleteOnFail}

    err := vail.PrintTestRestoreCsvHeader()
    if err != nil {
        return fmt.Errorf("failed printing header %v\n", err)
    }

    err = svc.ListObjectsV2Pages(
        &s3.ListObjectsV2Input{
            Bucket: aws.String(bucket),
            Prefix: aws.String(prefix),
            MaxKeys: aws.Int64(100)},
        vail.HandleTestByteRestore)
    return err
}

func doRestoreObject(svc *s3.S3, bucket string, key string) error {
    _, err := svc.RestoreObject(
        &s3.RestoreObjectInput{
            Bucket: aws.String(bucket),
            Key:  aws.String(key)})
    if err == nil {
        fmt.Printf("Restore requested: %s %s\n", key, time.Now().Format(time.RFC3339))
    } else {
        fmt.Printf("Restore request failed: %s %v\n", key, err)
    }
    return err
}

func restoreObject(svc *s3.S3, args *Arguments) error {
    if len(args.Key) > 0 {
        return doRestoreObject(svc, args.Bucket, args.Key)
    }
    if len(args.Prefix) > 0 {
        keyList, err := doBucketInventory(svc, args.Bucket, args.Prefix)
        if err != nil {
            return fmt.Errorf("failed getting object list %v\n", err)
        }
        if len(keyList) == 0 {
            return fmt.Errorf("no objects match bucket %s and prefix %s,%v\n",
                args.Bucket, args.Prefix, err)
        }
        atLeastOneGood := false
        for _, key := range keyList {
            err = doRestoreObject(svc, args.Bucket, *key.Key)
            if err == nil {
                atLeastOneGood = true
            }
        }
        if atLeastOneGood {
            // continue, some restoreObjects have succeeded
            return nil
        }
        return fmt.Errorf("No successful restore commands")
    }
    return fmt.Errorf("Must specify either key or prefix")
}

func headObject(svc *s3.S3, args *Arguments) error {
    restoreResponse, err := svc.HeadObject(
        &s3.HeadObjectInput{
            Bucket: aws.String(args.Bucket),
            Key:  aws.String(args.Key)})
    if err == nil {
        fmt.Printf("Restore request: %s\n", *restoreResponse.Restore)
    }
    return err
}

func getObjectByte(svc *s3.S3, args *Arguments) error {
    success, err := testGetObject(svc, args.Bucket, args.Key)
    if err != nil {
        return fmt.Errorf("could not issue test restore %v\n", err)
    }
    if success {
        fmt.Printf("SUCCESS test restoring %s\n", args.Key)
    } else {
        fmt.Printf("FAILED test restoring %s\n", args.Key)
    }
    return nil
}

func deleteObject(svc *s3.S3, args *Arguments) error {
    err := doDeleteObject(svc, args.Bucket, args.Key)

    if err != nil {
        return fmt.Errorf("could not issue deleteObject %v\n", err)
    }
    return nil
}

func getObject(svc *s3.S3, args *Arguments) error {
    // single object if key is defined
    if len(args.Key) > 0 {
        return doGetObject(svc, args.Bucket, args.Key)
    }

    // all objects in bucket matching prefix
    if len(args.Prefix) > 0 {
        keyList, err := doBucketInventory(svc, args.Bucket, args.Prefix)
        if err != nil {
            return fmt.Errorf("failed getting object list %v\n", err)
        }
        if len(keyList) == 0 {
            return fmt.Errorf("no objects match bucket %s and prefix %s,%v\n",
                args.Bucket, args.Prefix, err)
        }
        var wg sync.WaitGroup
        for _, key := range keyList {
            wg.Add(1)

            go func(name string) {
                err := doGetObject(svc, args.Bucket, name)
                if err != nil {
                    errorDescription := fmt.Sprintf("failed get-object for  '%s'%v\n", *key.Key, err)
                    log.Printf(errorDescription)
                    return
                }
                wg.Done()
            }(*key.Key)
        }
        wg.Wait()
    }
    return nil
}

func doGetObject(svc *s3.S3,  bucket string, key string) error {
    requestInput := &s3.GetObjectInput{
        Bucket: aws.String(bucket),
        Key:  aws.String(key),
    }

    getObjectRequest, getObjectResponse := svc.GetObjectRequest(requestInput)
    err := getObjectRequest.Send()
    if err != nil {
        return fmt.Errorf("falied to retrieve %s for bucket %s, %v\n",
            key, bucket, err)
    }

    // Get the last of the key
    fileName := path.Base(key)

    // Open the file to write.
    file, fileErr := os.Create(fileName)
    if fileErr != nil {
        return fileErr
    }
    defer file.Close()

    // Copy the request stream to the file.
    _, err = io.Copy(file, getObjectResponse.Body)
    if err != nil {
        return fmt.Errorf("falied to write object %s, %v\n",
            key, err)
    }
    fmt.Printf("Restored: %s\n", fileName)
    return nil
}

func testGetObject(svc *s3.S3,  bucket string, key string) (bool, error) {
    requestInput := &s3.GetObjectInput{
        Bucket: aws.String(bucket),
        Key:  aws.String(key),
        Range: aws.String("bytes=0-1"),
    }
    getObjectRequest, getObjectResponse := svc.GetObjectRequest(requestInput)
    err := getObjectRequest.Send()
    if err != nil {
        return false, fmt.Errorf("falied to retrieve %s from bucket %s, %v\n",
            key, bucket, err)
    }

    // ascertain that data is there
    var b bytes.Buffer
    w := bufio.NewWriter(&b)

    _, err = io.Copy(w, getObjectResponse.Body)
    // return false on error (expected); return error only if communication failed
    return err == nil, nil
}

func doDeleteObject(svc *s3.S3,  bucket string, key string) error {
    requestInput := &s3.DeleteObjectInput{
        Bucket: aws.String(bucket),
        Key:  aws.String(key),
    }
    _, err := svc.DeleteObject(requestInput)
    if err != nil {
        return fmt.Errorf("falied to delete %s from bucket %s, %v\n",
            key, bucket, err)
    }
    return nil
}

func restoreFromGlacier(svc *s3.S3, args *Arguments) error {
    // issue restore
    err := restoreObject(svc, args)
    if err != nil {
        return fmt.Errorf("failed restore request %v\n", err)
    }

    // wait until restore completes
    err = waitOnHead(svc, args)
    if err != nil {
        return fmt.Errorf("failed to restore %s, %v\n", args.Key, err)
    }

    // download if requested
    if args.Download == true {
        err = getObject(svc, args)
        if err != nil {
            return fmt.Errorf("failed to download %s, %v\n", args.Key, err)
        }
    }
    return nil
}

const maxInterval = 89
func waitOnHead(svc *s3.S3, args *Arguments) error {
    // single object if key is defined
    if len(args.Key) > 0 {
        fmt.Printf("Watching: %s %s\n", args.Key, time.Now().Format(time.RFC3339))
        err := doWaitOnHead(svc, args.Bucket, args.Key, 0, 1)
        if err != nil {
            return fmt.Errorf("failed waiting for  '%s'%v\n", args.Key, err)
        }
        fmt.Printf("Ready for download: %s %s\n", args.Key, time.Now().Format(time.RFC3339))
    }

    // all objects in bucket matching prefix
    if len(args.Prefix) > 0 {
        keyList, err := doBucketInventory(svc, args.Bucket, args.Prefix)
        if err != nil {
            return fmt.Errorf("failed getting object list %v\n", err)
        }
        if len(keyList) == 0 {
            return fmt.Errorf("no objects match bucket %s and prefix %s,%v\n",
                args.Bucket, args.Prefix, err)
        }
        atLeastOneGood := false
        var wg sync.WaitGroup

        for _, key := range keyList {
            wg.Add(1)

            go func(name string) {
                fmt.Printf("Watching: %s %s\n", name, time.Now().Format(time.RFC3339))
                err := doWaitOnHead(svc, args.Bucket, name, 0, 1)
                if err != nil {
                    errorDescription := fmt.Sprintf("failed waiting for  '%s'%v\n", name, err)
                    log.Printf(errorDescription)
                    return
                } else {
                    atLeastOneGood = true
                }
                wg.Done()
                fmt.Printf("Ready for download: %s %s\n", name, time.Now().Format(time.RFC3339))
            }(*key.Key)
        }
        wg.Wait()
        if !atLeastOneGood {
            return fmt.Errorf("no matching objects ready for restoration")
        }
    }
    return nil
}

func doWaitOnHead(svc *s3.S3, bucket string, key string, fib1 int, fib2 int) error {
    result, err := svc.HeadObject(
        &s3.HeadObjectInput{
            Bucket: aws.String(bucket),
            Key:  aws.String(key)})
    if err != nil {
        return fmt.Errorf("Head object failed: %v\n", err)
    }
    if *result.Restore != "ongoing-request=\"true\""  {
        return nil
    }
    // fibonacci up to maxInterval
    fib3 := fib1 + fib2
    if fib3 > maxInterval {
        fib3 = maxInterval
    }
    time.Sleep(time.Duration(fib3) * time.Second)
    // fmt.Printf(".")

    return doWaitOnHead(svc, bucket, key, fib2, fib3)
}
