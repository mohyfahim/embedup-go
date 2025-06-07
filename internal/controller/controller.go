package controller

import (
	"context"
	ApiClient "embedup-go/internal/apiclient"
	"embedup-go/internal/cstmerr"
	"embedup-go/internal/dbclient"
	SharedModels "embedup-go/internal/shared"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func FetchAndProcessContentUpdates(apiClientInstance *ApiClient.APIClient,
	dbConnection dbclient.DBClient,
	updater *SharedModels.Updater) error {
	params := SharedModels.ContentUpdateRequestParams{
		From:   updater.LastFromTimeStamp,
		Size:   50,
		Offset: 0,
	}

	response, processedItems, err := apiClientInstance.FetchContentUpdates(params)
	if err != nil {
		log.Printf("Failed to fetch content updates: %v", err)
		return err
	}

	if response == nil {
		log.Printf("No response received from content updates fetch.")
		return fmt.Errorf("nil response from FetchContentUpdates")
	}

	log.Printf("Fetched %d items, %d remaining in total on server.", len(processedItems), response.Count)

	for _, item := range processedItems {
		ProcessContentItem(item, dbConnection, apiClientInstance)
		//TODO: handle error in processing item
		if item.UpdatedAt > updater.LastFromTimeStamp {
			updater.LastFromTimeStamp = item.UpdatedAt
		}
	}

	//TODO: uncomment
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
	// defer cancel()
	// err = dbConnection.Save(ctx, &updater)
	// if err != nil {
	// 	log.Printf("Error on Updating lastFromTimestamp: %v", err)
	// 	return err
	// }

	return nil

}

// ProcessContentItem is a placeholder for the function that acts on parsed content.
func ProcessContentItem(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiClient *ApiClient.APIClient) {
	log.Printf("Processing item ID: %d, Type: %s, Enabled: %t", content.ID, content.Type, content.Enable)
	// Add your logic here based on content.Type and content.Details

	switch v := content.Details.(type) {
	case SharedModels.LocalAdvertisementSchema:
		log.Printf("Local Advertisement: FileLink: %s, SkipDuration: %d", v.FileLink, v.SkipDuration)
		ProcessLocalAdvertisement(content, dbConnection, apiClient)
		// case SharedModels.LocalSliderSchema:
	// 	log.Printf("Local Slider: ImageURL: %s, Num LocalTabIDs: %d", v.ImageURL, len(v.LocalTabIDs))
	// 	// Perform actions for LocalSliderContent
	// // Add cases for other types
	default:
		log.Printf("Cannot perform specific action for type %T", v)
	}
}

func ProcessLocalAdvertisement(
	content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiclient *ApiClient.APIClient) error {

	contentBasePath := os.Getenv("PODBOX_UPDATE_CONTENT_BASE_PATH")
	if contentBasePath == "" {
		contentBasePath = "/mnt/sdcard/assets/"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
	defer cancel()
	localAdvertisement := SharedModels.Advertisement{}
	localAdvertisementLink := SharedModels.AdvertisementLink{}
	localAdvertisement.ContentId = content.ID
	if content.Enable {
		detail := content.Details.(SharedModels.LocalAdvertisementSchema)
		// Download filelink to destination
		destinationPath := filepath.Join(contentBasePath, "videos", "ads")
		log.Printf("destination path for download file : %s \n", destinationPath)
		err := SharedModels.CheckAndCreateDir(destinationPath)
		if err != nil {
			log.Printf("Error in creating path %s: %v", destinationPath, err)
		}
		parts := strings.Split(detail.FileLink, "/")
		var podspaceHash string

		if len(parts) > 0 {
			podspaceHash = parts[len(parts)-1] + ".mp4"
		} else {
			return cstmerr.NewLinkParseError(fmt.Sprintf("failed to parse %s", detail.FileLink))
		}
		destinationFile := filepath.Join(destinationPath, podspaceHash)
		log.Printf("destination file: %s", destinationFile)
		var retryCount int = 0
		for {
			err := apiclient.DownloadFile(detail.FileLink, destinationFile)
			if err == nil {
				break
			}
			if retryCount == 3 {
				return cstmerr.NewRetryError("retry reached")
			}
			retryCount++
		}

		localAdvertisement.SkipDuration = int32(detail.SkipDuration)
		localAdvertisement.Synced = false
		localAdvertisementLink.LinkType = "MP4"
		_, err = SharedModels.CalculateMD5(destinationFile, 1025)
		if err != nil {
			log.Printf("error in calculating hash %v", err)
		}
		localAdvertisementLink.FileHash = "string(hash)"
		localAdvertisementLink.PlayLink = destinationFile
		localAdvertisementLink.OriginalLink = detail.FileLink
		localAdvertisement.Link = localAdvertisementLink
		dbConnection.Save(ctx, &localAdvertisement)
	} else {
		dbConnection.Delete(ctx, &localAdvertisement)
	}
	return nil
}
