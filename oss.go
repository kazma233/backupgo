package main

import (
	"backupgo/config"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var ErrCoolDown = errors.New("fast upload cool down")

type (
	NamedBucket struct {
		Name   string
		Bucket *oss.Bucket
	}

	OssClient struct {
		slowBucket      *NamedBucket
		fastBucket      *NamedBucket
		lastSuccessTime time.Time
	}

	UploadNoticeFunc func(string)
)

func CreateOSSClient(config config.OssConfig) *OssClient {
	ossClient := &OssClient{
		slowBucket: must(getBucket(
			"SLOW",
			config.Endpoint,
			config.AccessKey,
			config.AccessKeySecret,
			config.BucketName)), // slowBucket must not nil
		fastBucket: getBucket(
			"FAST",
			config.FastEndpoint,
			config.AccessKey,
			config.AccessKeySecret,
			config.BucketName,
		),
	}

	log.Printf("oss client init done: %v", ossClient)

	return ossClient
}

func (oc *OssClient) Upload(objKey, filePath string, noticeFunc UploadNoticeFunc) (err error) {
	if oc.slowBucket == nil && oc.fastBucket == nil {
		return errors.New("client not init")
	}

	err = oc.upload(oc.slowBucket, objKey, filePath, noticeFunc)
	if err == nil {
		return
	}

	if !oc.canUseFastBucket() {
		noticeFunc("fast bucket in 3-day cooldown")
		return ErrCoolDown
	}

	err = oc.upload(oc.fastBucket, objKey, filePath, noticeFunc)
	if err == nil {
		return
	}

	return
}

func (oc *OssClient) upload(bucket *NamedBucket, objKey, filePath string, noticeFunc UploadNoticeFunc) error {
	if bucket == nil || bucket.Bucket == nil {
		return fmt.Errorf("bucket %s not init", bucket.Name)
	}

	noticeFunc(fmt.Sprintf("use 【%s】 bucket uploading", bucket.Name))
	err := bucket.Bucket.PutObjectFromFile(objKey, filePath)
	if err != nil {
		noticeFunc(fmt.Sprintf("use 【%s】 bucket upload failed, error: %v", bucket.Name, err))
		return err
	}

	noticeFunc(fmt.Sprintf("use 【%s】 bucket upload success", bucket.Name))
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
	if oc.slowBucket == nil || oc.slowBucket.Bucket == nil {
		return "", errors.New("bucket not init")
	}

	return oc.slowBucket.Bucket.SignURL(objKey, oss.HTTPGet, 60*60*24*1)
}

func (oc *OssClient) GetSlowClient() *oss.Bucket {
	return oc.slowBucket.Bucket
}

func must[T any](obj T) T {
	if isNil(obj) {
		panic(errors.New("obj is nil"))
	}

	return obj
}

func isNil[T any](obj T) bool {
	v := reflect.ValueOf(obj)
	kind := v.Kind()
	return canBeNil(kind) && v.IsNil()
}

func canBeNil(kind reflect.Kind) bool {
	return kind == reflect.Pointer ||
		kind == reflect.Interface ||
		kind == reflect.Slice ||
		kind == reflect.Map ||
		kind == reflect.Chan ||
		kind == reflect.Func
}

func getBucket(customName, endpoint, ak, aks, buckatName string) *NamedBucket {
	if endpoint == "" || ak == "" || aks == "" || buckatName == "" {
		return nil
	}

	client, err := oss.New(endpoint, ak, aks, oss.Timeout(10, 60*60*3))
	if err != nil {
		panic(err)
	}

	bucket, err := client.Bucket(buckatName)
	if err != nil {
		panic(err)
	}

	return &NamedBucket{
		Name:   customName,
		Bucket: bucket,
	}
}
