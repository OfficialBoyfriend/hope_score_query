package utils

import (
	"context"
	"fmt"
	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/sms/bytes"
	"github.com/qiniu/api.v7/v7/storage"
	"os"
)

var (
	accessKey = os.Getenv("score_query_qiniu_access_key")
	secretKey = os.Getenv("score_query_qiniu_secret_key")
)

func UploadFile(bucket, key string, data []byte) (string, error) {
	putPolicy := storage.PutPolicy{
		Scope:               fmt.Sprintf("%v:%v", bucket, key),
	}
	mac := qbox.NewMac(accessKey, secretKey)
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = &storage.ZoneHuadong
	// 是否使用https域名
	cfg.UseHTTPS = false
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	// 构建表单上传的对象
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	// 数据
	fileData := bytes.NewReader(data)
	err := formUploader.Put(context.Background(), &ret, upToken, key, fileData, int64(len(data)), nil)
	if err != nil {
		return "", err
	}
	return ret.Hash, nil
}