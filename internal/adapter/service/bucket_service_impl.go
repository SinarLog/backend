package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"

	"cloud.google.com/go/storage"
	"sinarlog.com/pkg/bucket"
)

type bucketService struct {
	bkt *bucket.Bucket
}

func NewBucketService(bkt *bucket.Bucket) *bucketService {
	return &bucketService{bkt}
}

func (s *bucketService) upload(ctx context.Context, id string, prefix string, file multipart.File) (string, error) {
	path := prefix + fmt.Sprintf("/%s", id)

	obj := s.bkt.Handler.Object(path)
	// Optional: set a generation-match precondition to avoid potential race
	// conditions and data corruptions. The request to upload is aborted if the
	// object's generation number does not match your precondition.
	// For an object that does not yet exist, set the DoesNotExist precondition.
	// obj = obj.If(storage.Conditions{DoesNotExist: true})

	// Write file
	writer := obj.NewWriter(ctx)
	if _, err := io.Copy(writer, file); err != nil {
		writer.Close()
		return "", err
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	// BUG: Probably bugged here
	// Make the file public
	acl := obj.ACL()
	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", fmt.Errorf("ACLHandleSet: %w", err)
	}

	// Get link for accessed
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return "", fmt.Errorf("object.Attrs: %w", err)
	}
	link := fmt.Sprintf(s.bkt.PublicLinkTemplate, attrs.Bucket, url.QueryEscape(attrs.Name))

	return link, nil
}

func (s *bucketService) CreateAvatar(ctx context.Context, employeeId string, file multipart.File) (string, error) {
	return s.upload(ctx, employeeId, s.bkt.AvatarPath, file)
}

func (s *bucketService) CreateLeaveAttachment(ctx context.Context, leaveId string, file multipart.File) (string, error) {
	return s.upload(ctx, leaveId, s.bkt.LeaveAttachmentPath, file)
}

func (s *bucketService) delete(ctx context.Context, id, prefix string) error {
	path := prefix + fmt.Sprintf("/%s", id)

	obj := s.bkt.Handler.Object(path)

	// Optional: set a generation-match precondition to avoid potential race
	// conditions and data corruptions. The request to delete the file is aborted
	// if the object's generation number does not match your precondition.
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return fmt.Errorf("object.Attrs: %w", err)
	}
	obj = obj.If(storage.Conditions{GenerationMatch: attrs.Generation})

	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("Object(%q).Delete: %w", path, err)
	}

	return nil
}

func (s *bucketService) DeleteAvatar(ctx context.Context, employeeId string) error {
	return s.delete(ctx, employeeId, s.bkt.AvatarPath)
}

func (s *bucketService) DeleteLeaveAttachment(ctx context.Context, leaveId string) error {
	return s.delete(ctx, leaveId, s.bkt.LeaveAttachmentPath)
}
