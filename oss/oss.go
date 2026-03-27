package oss

import (
	"backupgo/config"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

var ErrCoolDown = errors.New("fast upload cool down")

type (
	OssClient struct {
		client          *oss.Client
		bucketName      string
		fastEndpoint    string
		fastClient      *oss.Client
		lastSuccessTime time.Time
	}

	UploadNoticeFunc func(string)
)

func CreateOSSClient(cfg config.OssConfig) *OssClient {
	client := oss.NewClient(oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.AccessKeySecret, "")).
		WithEndpoint(cfg.Endpoint))

	oc := &OssClient{
		client:       client,
		bucketName:   cfg.BucketName,
		fastEndpoint: cfg.FastEndpoint,
	}

	if cfg.FastEndpoint != "" {
		oc.fastClient = oss.NewClient(oss.LoadDefaultConfig().
			WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.AccessKeySecret, "")).
			WithEndpoint(cfg.FastEndpoint))
	}

	log.Printf("oss client init done: bucket=%s, slow=%s, fast=%s", cfg.BucketName, cfg.Endpoint, cfg.FastEndpoint)

	return oc
}

func (oc *OssClient) Upload(objKey, filePath string, noticeFunc UploadNoticeFunc) (err error) {
	if oc.client == nil {
		return errors.New("client not init")
	}

	err = oc.upload(oc.client, objKey, filePath, noticeFunc)
	if err == nil {
		return
	}

	if !oc.canUseFastBucket() || oc.fastClient == nil {
		noticeFunc("fast bucket in 3-day cooldown")
		return ErrCoolDown
	}

	err = oc.upload(oc.fastClient, objKey, filePath, noticeFunc)
	if err == nil {
		return
	}

	return
}

func (oc *OssClient) upload(client *oss.Client, objKey, filePath string, noticeFunc UploadNoticeFunc) error {
	noticeFunc(fmt.Sprintf("uploading to %s", objKey))
	_, err := client.PutObjectFromFile(context.Background(), &oss.PutObjectRequest{
		Bucket: oss.Ptr(oc.bucketName),
		Key:    oss.Ptr(objKey),
	}, filePath)
	if err != nil {
		noticeFunc(fmt.Sprintf("upload failed: %v", err))
		return err
	}

	noticeFunc(fmt.Sprintf("upload success: %s", objKey))
	oc.setLastSuccessTime()

	return nil
}

func (oc *OssClient) HasError(err error) bool {
	return err != nil && err != ErrCoolDown
}

func (oc *OssClient) HasCoolDownError(err error) bool {
	return err == ErrCoolDown
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
