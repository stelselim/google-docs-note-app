package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"sync"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var (
	GoogleDriveService *drive.Service
	initDriveOnce      sync.Once
	initDriveErr       error
)

func ConnectGoogleDrive(ctx context.Context) (*drive.Service, error) {
	initDriveOnce.Do(func() {
		// Create Drive service
		driveService, err := drive.NewService(ctx, option.WithScopes(drive.DriveScope))
		if err != nil {
			initDriveErr = fmt.Errorf("unable to create Drive client: %v", err)
		}
		fmt.Println("Successfully connected to Google Drive")

		GoogleDriveService = driveService
	})
	return GoogleDriveService, initDriveErr
}

func UploadImageToDrive(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	defer file.Close()

	f := &drive.File{
		Name:     fileHeader.Filename,
		MimeType: fileHeader.Header.Get("Content-Type"),
	}

	uploadedFile, uploadErr := GoogleDriveService.Files.Create(f).Media(file).Do()
	if uploadErr != nil {
		return "", uploadErr
	}
	// Set image permission to public to add Google Docs
	_, permissionSetErr := GoogleDriveService.Permissions.Create(uploadedFile.Id, &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}).Do()

	if permissionSetErr != nil {
		return "", permissionSetErr
	}

	return uploadedFile.Id, nil
}

func GetFileUrlById(fileId string) string {
	return "https://drive.google.com/uc?id=" + fileId + "&sz=w1000"
}

func DeleteImageFromDrive(fileId string) error {
	deleteErr := GoogleDriveService.Files.Delete(fileId).Do()
	return deleteErr
}
