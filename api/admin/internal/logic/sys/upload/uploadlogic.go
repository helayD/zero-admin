package upload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
)

const (
	defaultUploadMaxSizeMB = 20
	uploadFormField        = "file"
)

// UploadLogic 文件上传
/*
Author: LiuFeiHua
Date: 2024/5/27 9:23
*/
type UploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewUploadLogic(r *http.Request, ctx context.Context, svcCtx *svc.ServiceContext) *UploadLogic {
	return &UploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Upload 文件上传
func (l *UploadLogic) Upload() (resp *types.UploadResp, err error) {
	maxSizeMB := effectiveMaxSizeMB(l.svcCtx.Config.SystemConfig.OSS.MaxSizeMB)
	l.r.Body = http.MaxBytesReader(nil, l.r.Body, maxSizeMB*1024*1024)

	reader, err := l.r.MultipartReader()
	if err != nil {
		logc.Errorf(l.ctx, "解析上传请求失败,异常:%s", err.Error())
		return nil, errors.New("上传请求格式不正确")
	}

	for {
		part, partErr := reader.NextPart()
		if errors.Is(partErr, io.EOF) {
			break
		}
		if partErr != nil {
			logc.Errorf(l.ctx, "读取上传文件失败,异常:%s", partErr.Error())
			return nil, partErr
		}

		if part.FormName() != uploadFormField || part.FileName() == "" {
			_ = part.Close()
			continue
		}

		fileURL, uploadErr := l.uploadToOSS(part)
		_ = part.Close()
		if uploadErr != nil {
			return nil, uploadErr
		}

		logx.WithContext(l.ctx).Infof("文件上传 OSS 成功: %s", part.FileName())
		return &types.UploadResp{
			Code:    "000000",
			Message: "上传文件成功",
			Data:    fileURL,
		}, nil
	}

	return nil, errors.New("请选择要上传的文件")
}

func (l *UploadLogic) uploadToOSS(part *multipart.Part) (string, error) {
	ossConfig := l.svcCtx.Config.SystemConfig.OSS
	endpoint := strings.TrimSpace(ossConfig.Endpoint)
	accessKeyID := strings.TrimSpace(ossConfig.AccessKeyID)
	accessKeySecret := strings.TrimSpace(ossConfig.AccessKeySecret)
	bucketName := strings.TrimSpace(ossConfig.BucketName)

	if endpoint == "" || strings.EqualFold(endpoint, "none") || accessKeyID == "" || accessKeySecret == "" || bucketName == "" {
		return "", errors.New("OSS配置不完整")
	}

	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		logc.Errorf(l.ctx, "创建 OSS 客户端失败,异常:%s", err.Error())
		return "", err
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		logc.Errorf(l.ctx, "获取 OSS Bucket 失败,异常:%s", err.Error())
		return "", err
	}

	objectKey := buildObjectKey(part.FileName())
	options := make([]oss.Option, 0, 1)
	if contentType := resolveContentType(part); contentType != "" {
		options = append(options, oss.ContentType(contentType))
	}

	if err = bucket.PutObject(objectKey, part, options...); err != nil {
		logc.Errorf(l.ctx, "上传文件到 OSS 失败,文件:%s,对象:%s,异常:%s", part.FileName(), objectKey, err.Error())
		if strings.Contains(err.Error(), "request body too large") {
			return "", fmt.Errorf("上传文件不能超过%dMB", effectiveMaxSizeMB(ossConfig.MaxSizeMB))
		}
		return "", err
	}

	return buildObjectURL(ossConfig.URL, endpoint, bucketName, objectKey), nil
}

func buildObjectKey(filename string) string {
	cleanName := strings.ReplaceAll(filename, "\\", "/")
	ext := strings.ToLower(path.Ext(path.Base(cleanName)))
	return fmt.Sprintf("uploads/%s/%s%s", time.Now().Format("20060102"), uuid.NewString(), ext)
}

func effectiveMaxSizeMB(maxSizeMB int64) int64 {
	if maxSizeMB <= 0 {
		return defaultUploadMaxSizeMB
	}
	return maxSizeMB
}

func resolveContentType(part *multipart.Part) string {
	if contentType := strings.TrimSpace(part.Header.Get("Content-Type")); contentType != "" {
		return contentType
	}

	ext := strings.ToLower(path.Ext(part.FileName()))
	if ext == "" {
		return "application/octet-stream"
	}
	if contentType := mime.TypeByExtension(ext); contentType != "" {
		return contentType
	}
	return "application/octet-stream"
}

func buildObjectURL(baseURL, endpoint, bucketName, objectKey string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL != "" {
		return strings.TrimRight(baseURL, "/") + "/" + objectKey
	}

	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	return fmt.Sprintf("https://%s.%s/%s", bucketName, endpoint, objectKey)
}
