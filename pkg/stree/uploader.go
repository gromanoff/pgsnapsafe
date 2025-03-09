package stree

import (
	"context"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

// UploadFileToS3 uploads a local file to S3 and returns the URL
func UploadFileToS3(stree *s3.Client, bucketName string, filePath string, folder string) (string, error) {
	log.Println("🚀 Starting file upload to S3", "filePath", filePath, "folder", folder)

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("❌ Error opening file", "filePath", filePath, "error", err)
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	log.Println("📂 File successfully opened", "filePath", filePath)

	// Get file information
	fileInfo, err := file.Stat()
	if err != nil {
		log.Println("❌ Error getting file information", "filePath", filePath, "error", err)
		return "", fmt.Errorf("failed to get file information for %s: %w", filePath, err)
	}

	log.Println("📏 File size", "size", fileInfo.Size(), "filePath", filePath)

	// Generate unique file name
	uniqueFileName := fmt.Sprintf("%s_%s%s",
		time.Now().Format("20060102150405"),
		uuid.New().String(),
		strings.ToLower(filepath.Ext(filePath)),
	)
	objectKey := fmt.Sprintf("%s/%s", folder, uniqueFileName)

	log.Println("🔑 Generated unique S3 key", "objectKey", objectKey)

	// Determine MIME type
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	log.Println("📝 Determined MIME type", "contentType", contentType)

	log.Println("🪣 Using bucket", "bucket", bucketName)

	// Create PutObjectInput
	input := &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(objectKey),
		Body:          file,
		ContentType:   aws.String(contentType),
		ACL:           types.ObjectCannedACLPublicRead, // Make the file public
		ContentLength: aws.Int64(fileInfo.Size()),
	}

	log.Println("☁️ Sending file to S3", "bucket", bucketName, "objectKey", objectKey)

	// Upload file to S3
	_, err = stree.PutObject(context.TODO(), input)
	if err != nil {
		log.Println("❌ Error uploading file to S3", "bucket", bucketName, "objectKey", objectKey, "error", err)
		return "", fmt.Errorf("error uploading file to S3: %w", err)
	}

	log.Println("✅ File successfully uploaded to S3", "bucket", bucketName, "objectKey", objectKey)

	return objectKey, nil
}

// convertMapToAWSMetadata converts map[string]string to map[string]*string
func convertMapToAWSMetadata(metadata map[string]string) map[string]*string {
	converted := make(map[string]*string)
	for k, v := range metadata {
		val := v // create a copy of the value for correct pointer
		converted[k] = &val
	}
	return converted
}

// ListFilesInS3Directory gets a list of files from the specified directory in S3
func ListFilesInS3Directory(stree *s3.Client, bucketName string, folder string) ([]string, error) {

	// Form prefix for file search (directory + "/")
	prefix := fmt.Sprintf("%s/", folder)

	// Request to get list of objects with specified prefix
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	}

	// Execute request to S3
	output, err := stree.ListObjectsV2(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("error getting list of files from S3: %w", err)
	}

	// Form list of file names
	var files []string
	for _, item := range output.Contents {
		files = append(files, *item.Key)
	}

	return files, nil
}

// DeleteFileFromS3 deletes a file from S3 by its path (objectKey)
func DeleteFileFromS3(stree *s3.Client, bucketName string, filePath string) error {

	// Form delete request
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filePath),
	}

	// Execute deletion
	_, err := stree.DeleteObject(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error deleting file %s from S3: %w", filePath, err)
	}

	fmt.Printf("✅ File %s successfully deleted from S3\n", filePath)
	return nil
}
