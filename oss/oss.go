package oss

import (
	"backupgo/config"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

var ErrCoolDown = errors.New("加速上传冷却中")

type (
	BucketType string

	UploadResult struct {
		Bucket string
		Key    string
		Mode   BucketType
	}

	OssClient struct {
		bucketName      string
		client          *oss.Client
		fastClient      *oss.Client
		lastSuccessTime time.Time
	}
)

var (
	NORMAL BucketType = "NORMAL"
	FAST   BucketType = "FAST"
)

func CreateOSSClient(cfg config.OssConfig) *OssClient {
	client := oss.NewClient(newClientConfig(cfg, false))
	fastClient := oss.NewClient(newClientConfig(cfg, true))

	oc := &OssClient{
		client:     client,
		fastClient: fastClient,
		bucketName: strings.TrimSpace(cfg.BucketName),
	}

	log.Printf("oss client init done: bucket %s", oc.bucketName)

	return oc
}

func newClientConfig(cfg config.OssConfig, useFastEndpoint bool) *oss.Config {
	clientConfig := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.AccessKeySecret))
	region := strings.TrimSpace(cfg.Region)

	if region != "" {
		clientConfig.WithRegion(region)
	}
	if useFastEndpoint {
		clientConfig.WithUseAccelerateEndpoint(true)
	}

	return clientConfig
}

func (oc *OssClient) BucketName() string {
	return oc.bucketName
}

func (oc *OssClient) Upload(objKey, filePath string) (UploadResult, error) {
	result := UploadResult{
		Bucket: oc.bucketName,
		Key:    objKey,
		Mode:   NORMAL,
	}

	err := upload(oc.client, oc.bucketName, objKey, filePath)
	if err == nil {
		oc.setLastSuccessTime()
		return result, nil
	}
	normalErr := err

	if !oc.canUseFastBucket() {
		return result, fmt.Errorf("普通上传失败: %v；%w", normalErr, ErrCoolDown)
	}

	result.Mode = FAST
	err = upload(oc.fastClient, oc.bucketName, objKey, filePath)
	if err == nil {
		oc.setLastSuccessTime()
		return result, nil
	}

	return result, fmt.Errorf("普通上传失败: %v；加速上传失败: %w", normalErr, err)
}

func upload(client *oss.Client, bucketName, objKey, filePath string) error {
	_, err := client.PutObjectFromFile(context.Background(), &oss.PutObjectRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objKey),
	}, filePath)

	return err
}

func (oc *OssClient) canUseFastBucket() bool {
	if oc.lastSuccessTime.IsZero() {
		return true
	}
	return time.Since(oc.lastSuccessTime) > 3*24*time.Hour
}

func (oc *OssClient) setLastSuccessTime() {
	oc.lastSuccessTime = time.Now()
}

func (oc *OssClient) TempVisitLink(objKey string) (string, error) {
	if oc.client == nil {
		return "", errors.New("client not init")
	}

	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(oc.bucketName),
		Key:    oss.Ptr(objKey),
	}
	result, err := oc.client.GetObject(context.Background(), request)
	if err != nil {
		return "", err
	}
	defer result.Body.Close()

	return "", nil
}

func (oc *OssClient) DeleteObjectsByPredicate(shouldDelete func(key string) bool) ([]string, error) {
	var keys []string

	p := oc.client.NewListObjectsV2Paginator(&oss.ListObjectsV2Request{
		Bucket: oss.Ptr(oc.bucketName),
	})

	for p.HasNext() {
		page, err := p.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		for _, obj := range page.Contents {
			key := oss.ToString(obj.Key)
			if shouldDelete(key) {
				keys = append(keys, key)
			}
		}
	}

	if len(keys) == 0 {
		return nil, nil
	}

	var deleteObjects []oss.DeleteObject
	for _, key := range keys {
		deleteObjects = append(deleteObjects, oss.DeleteObject{Key: oss.Ptr(key)})
	}

	result, err := oc.client.DeleteMultipleObjects(context.Background(), &oss.DeleteMultipleObjectsRequest{
		Bucket: oss.Ptr(oc.bucketName),
		Delete: &oss.Delete{Objects: deleteObjects},
	})
	if err != nil {
		return nil, err
	}

	var deleted []string
	for _, d := range result.DeletedObjects {
		deleted = append(deleted, oss.ToString(d.Key))
	}
	return deleted, nil
}
