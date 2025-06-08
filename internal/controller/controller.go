package controller

import (
	"context"
	ApiClient "embedup-go/internal/apiclient"
	"embedup-go/internal/cstmerr"
	"embedup-go/internal/dbclient"
	SharedModels "embedup-go/internal/shared"
	"encoding/hex"
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
func ProcessContentItem(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiClient *ApiClient.APIClient) error {
	log.Printf("Processing item ID: %d, Type: %s, Enabled: %t", content.ID, content.Type, content.Enable)

	switch v := content.Details.(type) {
	case SharedModels.LocalAdvertisementSchema:
		return ProcessLocalAdvertisement(content, dbConnection, apiClient)
	case SharedModels.LocalPageSchema:
		return ProcessLocalPage(content, dbConnection)
	case SharedModels.LocalTabSchema:
		return ProcessLocalTab(content, dbConnection)
	case SharedModels.LocalSliderSchema:
		return ProcessLocalSlider(content, dbConnection)
	// 	log.Printf("Local Slider: ImageURL: %s, Num LocalTabIDs: %d", v.ImageURL, len(v.LocalTabIDs))
	// 	// Perform actions for LocalSliderContent
	// // Add cases for other types
	default:
		log.Printf("Cannot perform specific action for type %T", v)
	}

	return nil
}

func ProcessLocalSlider(content SharedModels.ProcessedContentSchema, dbConnection dbclient.DBClient) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
	defer cancel()

	localSlider := SharedModels.Slider{}
	detail := content.Details.(SharedModels.LocalSliderSchema)

	localSlider.ContentId = content.ID
	localSlider.ButtonTitle = detail.ButtonTitle
	localSlider.Image.ImageURL = detail.ImageURL
	localSlider.Image.LogoImageUrl = detail.LogoImageURL
	localSlider.Image.MediumImageUrl = &detail.MediumImageURL
	localSlider.Image.SmallImageUrl = &detail.SmallImageURL
	localSlider.Link = detail.Link

	err := dbConnection.Save(ctx, &localSlider)
	if err != nil {
		return cstmerr.NewProcessError("failed to create slider", err)
	}
	if len(detail.LocalTabIDs) > 0 {
		tabs := make([]*SharedModels.Tab, len(detail.LocalTabIDs))
		for index, value := range detail.LocalTabIDs {
			tab := SharedModels.Tab{}
			tab.ContentId = int64(value)
			tabs[index] = &tab
		}

		err = dbConnection.CreateWithAssosiate(ctx, &localSlider, "Tabs", &tabs)
		if err != nil {
			return cstmerr.NewProcessError("failed to create assosiate tab page", err)
		}
	}
	return nil
}

func ProcessLocalTab(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
	defer cancel()

	localTab := SharedModels.Tab{}
	detail := content.Details.(SharedModels.LocalTabSchema)
	localTab.ContentId = content.ID
	localTab.Name = detail.Name
	localTab.Type = detail.Type

	err := dbConnection.Save(ctx, &localTab)
	if err != nil {
		return cstmerr.NewProcessError("failed to create tab", err)
	}
	if len(detail.LocalPageIDs) > 0 {
		pages := make([]*SharedModels.Page, len(detail.LocalPageIDs))

		for index, value := range detail.LocalPageIDs {
			page := SharedModels.Page{}
			page.ContentId = int64(value)
			pages[index] = &page
		}
		err := dbConnection.CreateWithAssosiate(ctx, &localTab, "Pages", &pages)
		if err != nil {
			return cstmerr.NewProcessError("failed to create assosiate tab page", err)
		}
	}

	return nil
}
func ProcessLocalPage(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
	defer cancel()
	localPage := SharedModels.Page{}
	detail := content.Details.(SharedModels.LocalPageSchema)
	localPage.ContentId = content.ID
	localPage.Name = &detail.Name
	localPage.Type = detail.Type
	err := dbConnection.Save(ctx, &localPage)
	if err != nil {
		return cstmerr.NewProcessError("failed to save Local Page", err)
	}
	return nil
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

		err = apiclient.DownloadFileWithRetry(detail.FileLink, destinationFile)

		if err != nil {
			log.Printf("error in downloading hash")
			return cstmerr.NewDownloadError(
				fmt.Sprintf("failed to download multiple times: %s", detail.FileLink))
		}

		localAdvertisement.SkipDuration = int32(detail.SkipDuration)
		localAdvertisement.Synced = false
		localAdvertisementLink.LinkType = "MP4"
		hash, err := SharedModels.CalculateMD5(destinationFile, 1025)
		if err != nil {
			log.Printf("error in calculating hash %v", err)
		}
		localAdvertisementLink.FileHash = hex.EncodeToString(hash)
		localAdvertisementLink.PlayLink = destinationFile
		localAdvertisementLink.OriginalLink = detail.FileLink
		localAdvertisement.Link = localAdvertisementLink
		dbConnection.Save(ctx, &localAdvertisement)
	} else {
		dbConnection.Delete(ctx, &localAdvertisement)
	}
	return nil
}
