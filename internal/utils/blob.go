package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cfg "backend/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func InitS3SDK(server_cfg *cfg.Config) {
	// var bucketName = server_cfg.R2_bucket_name
	fmt.Println("Init S3")
	var accessKeyId = server_cfg.R2_access_key_id
	var accessKeySecret = server_cfg.R2_access_key
	var r2_endpoint = server_cfg.R2_endpoint
	log.Printf("Endpoint: %s\n", r2_endpoint)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(r2_endpoint)
	})
	// bucket is empty
	// listObjectsOutput, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
	// 	Bucket: &bucketName,
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// for _, object := range listObjectsOutput.Contents {
	// 	obj, _ := json.MarshalIndent(object, "", "\t")
	// 	fmt.Println(string(obj))
	// }

	//  {
	//    "ChecksumAlgorithm": null,
	//    "ETag": "\"eb2b891dc67b81755d2b726d9110af16\"",
	//    "Key": "ferriswasm.png",
	//    "LastModified": "2022-05-18T17:20:21.67Z",
	//    "Owner": null,
	//    "Size": 87671,
	//    "StorageClass": "STANDARD"
	//  }

	listBucketsOutput, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatal(err)
	}

	for _, object := range listBucketsOutput.Buckets {
		obj, _ := json.MarshalIndent(object, "", "\t")
		fmt.Println(string(obj))
	}

	// {
	//     "CreationDate": "2022-05-18T17:19:59.645Z",
	//     "Name": "sdk-example"
	// }
}
