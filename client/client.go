package client

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"strconv"
	"time"
)

type VailClient struct {
	Client 		*s3.S3
	Csv  		*csv.Writer
	Bucket  	string
	Prefix  	string
	DeleteOnFail bool
}

func (vail *VailClient) PrintObjectsPage (resp *s3.ListObjectsV2Output, more bool) bool {
	_ = printObjectList(resp.Contents, vail.Csv)
	return *resp.IsTruncated
}

func printObjectList(objects  []*s3.Object, csv *csv.Writer) error {
	for _, object :=  range objects {
		var line = []string {*object.Key, strconv.FormatInt(*object.Size, 10), *object.StorageClass,  object.LastModified.Format(time.RFC822)}
		_ = csv.Write(line)	}
	return nil
}

func (vail *VailClient) HandleTestByteRestore(resp *s3.ListObjectsV2Output, more bool) bool {
	// let them know we're here
	fmt.Printf(".")
	for _, object := range resp.Contents {
		// Ignore glacier class
		class := *object.StorageClass
		if class == "GLACIER" {
			var line = []string{*object.Key,
				strconv.FormatBool(false), class, "", ""}
			_ = vail.Csv.Write(line)
			continue
		}

		success, err := testGetObject(vail.Client, vail.Bucket, *object.Key)
		errorString := ""
		deleteErrorString := ""
		deleted := ""
		if err != nil {
			errorString = fmt.Sprintf("ERR: %v", err)
		}
		if vail.DeleteOnFail && !success{
			err = doDeleteObject(vail.Client, vail.Bucket, *object.Key)
			if err != nil {
				deleteErrorString = fmt.Sprintf("ERR: %v", err)
			} else {
				deleted = "Deleted"
			}
		}
		var line = []string {*object.Key,
			strconv.FormatBool(success), deleted,  errorString, deleteErrorString}
		_ = vail.Csv.Write(line)
	}
	return *resp.IsTruncated
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

func (vail *VailClient) PrintTestRestoreCsvHeader() error {
	var line = []string {"Key","Restorable","Deleted","Error","Delete Error"}
	return vail.Csv.Write(line)
}

func (vail *VailClient) PrintListBucketsCsvHeader() error {
	var line = []string {"Name","Creation Date"}
	return vail.Csv.Write(line)
}

func (vail *VailClient) PrintBucketObjectsCsvHeader() error {
	var line = []string {"Key","Size","Storage Class","Creation Date"}
	return vail.Csv.Write(line)
}

func (vail *VailClient) PrintBucketList(buckets  []*s3.Bucket) error {
	for _, bucket :=  range buckets {
		var line = []string {*bucket.Name, bucket.CreationDate.Format(time.RFC822)}
		_ = vail.Csv.Write(line)
	}
	return nil
}

