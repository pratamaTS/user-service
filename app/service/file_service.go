package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"

	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/helpers"
	"harjonan.id/user-service/app/infra/r2"
	"harjonan.id/user-service/app/repository"
)

type FileService interface {
	Upload(ctx *gin.Context)
	Get(ctx *gin.Context)
}

type FileServiceImpl struct {
	r2   *r2.Client
	repo repository.ImageRepository
}

func NewFileService(r2c *r2.Client, repo repository.ImageRepository) *FileServiceImpl {
	return &FileServiceImpl{r2: r2c, repo: repo}
}

// POST /files/upload  (multipart/form-data)
// form:
// - file: (required)
// - key: (optional) -> kalau kosong, auto generate
func (s *FileServiceImpl) Upload(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		helpers.JsonErr[any](ctx, "missing file", http.StatusBadRequest, err)
		return
	}

	// optional input key
	key := strings.TrimSpace(ctx.PostForm("key"))
	if key == "" {
		ext := filepath.Ext(file.Filename)
		if ext == "" {
			ext = ".bin"
		}
		key = fmt.Sprintf("uploads/%s%s", helpers.GenerateUUID(), ext)
	}

	f, err := file.Open()
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to open file", http.StatusBadRequest, err)
		return
	}
	defer f.Close()

	// detect content-type (best effort)
	contentType := mime.TypeByExtension(filepath.Ext(file.Filename))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	putCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Print("Key: ", key)
	_, err = s.r2.S3.PutObject(putCtx, &s3.PutObjectInput{
		Bucket:      &s.r2.Bucket,
		Key:         &key,
		Body:        f,
		ContentType: &contentType,
	})
	if err != nil {
		helpers.JsonErr[any](ctx, "upload failed", http.StatusInternalServerError, err)
		return
	}

	if s.r2.PublicBaseURL == "" {
		helpers.JsonErr[any](ctx, "missing R2_PUBLIC_BASE_URL (url cannot be built)", http.StatusInternalServerError, errors.New("R2_PUBLIC_BASE_URL is empty"))
		return
	}

	publicURL := strings.TrimRight(s.r2.PublicBaseURL, "/") + "/" + key

	now := time.Now()
	img := dao.Image{
		UUID:         helpers.GenerateUUID(),
		Bucket:       s.r2.Bucket,
		Key:          key,
		URL:          publicURL,
		ContentType:  contentType,
		Size:         file.Size,
		CreatedAt:    now,
		CreatedAtStr: now.Format(time.RFC3339),
	}

	saved, err := s.repo.Save(&img)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to save image metadata", http.StatusInternalServerError, err)
		return
	}

	helpers.JsonOK(ctx, "success", saved)
}

// GET /files/:key  (proxy download)
// - streaming dari R2 -> client
func (s *FileServiceImpl) Get(ctx *gin.Context) {
	key := ctx.Param("key")
	key = strings.TrimPrefix(key, "/")
	if key == "" {
		helpers.JsonErr[any](ctx, "missing key", http.StatusBadRequest, errors.New("key required"))
		return
	}

	getCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	out, err := s.r2.S3.GetObject(getCtx, &s3.GetObjectInput{
		Bucket: &s.r2.Bucket,
		Key:    &key,
	})
	if err != nil {
		helpers.JsonErr[any](ctx, "not found", http.StatusNotFound, err)
		return
	}
	defer out.Body.Close()

	if out.ContentType != nil && *out.ContentType != "" {
		ctx.Header("Content-Type", *out.ContentType)
	}

	// optional: cache header (buat asset)
	ctx.Header("Cache-Control", "public, max-age=31536000, immutable")

	// stream body
	_, copyErr := io.Copy(ctx.Writer, out.Body)
	if copyErr != nil {
		// jangan JsonErr karena body sudah streaming
		return
	}
}
