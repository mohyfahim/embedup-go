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
	"time"
)

func DeleteAudio(filePath string) error {
	// Delete the file at the specified filePath
	contentBasePath := os.Getenv("PODBOX_UPDATE_CONTENT_BASE_PATH")
	if contentBasePath == "" {
		contentBasePath = "/mnt/sdcard/assets/"
	}
	dest := filepath.Join(contentBasePath, "audios", filePath)
	err := os.Remove(dest)
	if err != nil {
		log.Printf("Error deleting file %s: %v", dest, err)
		return cstmerr.NewFileDeleteError(fmt.Sprintf("failed to delete file: %s", dest), err)
	}
	return nil
}

func DeleteVideo(filePath string) error {
	// Delete the file at the specified filePath
	contentBasePath := os.Getenv("PODBOX_UPDATE_CONTENT_BASE_PATH")
	if contentBasePath == "" {
		contentBasePath = "/mnt/sdcard/assets/"
	}
	dest := filepath.Join(contentBasePath, "videos", filePath)
	err := os.Remove(dest)
	if err != nil {
		log.Printf("Error deleting file %s: %v", dest, err)
		return cstmerr.NewFileDeleteError(fmt.Sprintf("failed to delete file: %s", dest), err)
	}
	return nil
}

func DeleteImage(filePath string) error {
	// Delete the file at the specified filePath
	contentBasePath := os.Getenv("PODBOX_UPDATE_CONTENT_BASE_PATH")
	if contentBasePath == "" {
		contentBasePath = "/mnt/sdcard/assets/"
	}
	dest := filepath.Join(contentBasePath, "images", filePath)
	err := os.Remove(dest)
	if err != nil {
		log.Printf("Error deleting file %s: %v", dest, err)
		return cstmerr.NewFileDeleteError(fmt.Sprintf("failed to delete file: %s", dest), err)
	}
	return nil
}

func DownloadImage(apiclient *ApiClient.APIClient, url string, dir ...string) (string, string, error) {

	contentBasePath := os.Getenv("PODBOX_UPDATE_CONTENT_BASE_PATH")
	if contentBasePath == "" {
		contentBasePath = "/mnt/sdcard/assets/"
	}
	destinationPath := filepath.Join(append([]string{contentBasePath, "images"}, dir...)...)

	log.Printf("destination path for download file : %s \n", destinationPath)
	err := SharedModels.CheckAndCreateDir(destinationPath)
	if err != nil {
		log.Printf("Error in creating path %s: %v", destinationPath, err)
	}

	fileInformation, err := apiclient.GetFileInformation(url)

	if err != nil {
		fileInformation.MD5 = SharedModels.CalculateStringMD5(url)
	}

	fileNameWithPrefix := fileInformation.MD5 + ".jpg"

	destinationFile := filepath.Join(destinationPath, fileNameWithPrefix)
	log.Printf("destination file: %s", destinationFile)

	err = apiclient.DownloadFileWithRetry(url, destinationFile)

	if err != nil {
		log.Printf("error in downloading hash")
		return "", "", cstmerr.NewDownloadError(
			fmt.Sprintf("failed to download multiple times: %s", url))
	}

	return destinationFile, fileNameWithPrefix, nil
}

func DownloadVideo(apiclient *ApiClient.APIClient, url string, dir ...string) (string, string, error) {

	contentBasePath := os.Getenv("PODBOX_UPDATE_CONTENT_BASE_PATH")
	if contentBasePath == "" {
		contentBasePath = "/mnt/sdcard/assets/"
	}
	destinationPath := filepath.Join(append([]string{contentBasePath, "videos"}, dir...)...)

	log.Printf("destination path for download file : %s \n", destinationPath)
	err := SharedModels.CheckAndCreateDir(destinationPath)
	if err != nil {
		log.Printf("Error in creating path %s: %v", destinationPath, err)
	}

	fileInformation, err := apiclient.GetFileInformation(url)

	if err != nil {
		fileInformation.MD5 = SharedModels.CalculateStringMD5(url)
	}

	fileNameWithPrefix := fileInformation.MD5 + ".mp4"

	destinationFile := filepath.Join(destinationPath, fileNameWithPrefix)
	log.Printf("destination file: %s", destinationFile)

	err = apiclient.DownloadFileWithRetry(url, destinationFile)

	if err != nil {
		log.Printf("error in downloading hash")
		return "", "", cstmerr.NewDownloadError(
			fmt.Sprintf("failed to download multiple times: %s", url))
	}

	return destinationFile, fileNameWithPrefix, nil
}

func DownloadZippedVideo(apiclient *ApiClient.APIClient, url string, dir ...string) (string, string, error) {

	contentBasePath := os.Getenv("PODBOX_UPDATE_CONTENT_BASE_PATH")
	if contentBasePath == "" {
		contentBasePath = "/mnt/sdcard/assets/"
	}
	destinationPath := filepath.Join(append([]string{contentBasePath, "videos"}, dir...)...)

	log.Printf("destination path for download file : %s \n", destinationPath)
	err := SharedModels.CheckAndCreateDir(destinationPath)
	if err != nil {
		log.Printf("Error in creating path %s: %v", destinationPath, err)
	}

	fileInformation, err := apiclient.GetFileInformation(url)

	if err != nil {
		fileInformation.MD5 = SharedModels.CalculateStringMD5(url)
	}

	fileNameWithPrefix := fileInformation.MD5 + ".zip"

	destinationFile := filepath.Join(destinationPath, fileNameWithPrefix)
	log.Printf("destination file: %s", destinationFile)

	err = apiclient.DownloadFileWithRetry(url, destinationFile)

	if err != nil {
		log.Printf("error in downloading hash")
		return "", "", cstmerr.NewDownloadError(
			fmt.Sprintf("failed to download multiple times: %s", url))
	}
	destinationExtracted := filepath.Join(destinationPath, fileInformation.MD5)
	//TODO: enable this
	// err = SharedModels.UnzipFile(destinationFile, destinationExtracted)
	// if err != nil {
	// 	return "", "", err
	// }
	return destinationExtracted, fileNameWithPrefix, nil
}

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
		err := ProcessContentItem(item, dbConnection, apiClientInstance)
		if err != nil {
			return err
		}
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
	// case SharedModels.LocalPageSchema:
	// 	return ProcessLocalPage(content, dbConnection)
	// case SharedModels.LocalTabSchema:
	// 	return ProcessLocalTab(content, dbConnection)
	// case SharedModels.LocalSliderSchema:
	// 	return ProcessLocalSlider(content, dbConnection, apiClient)
	// case SharedModels.LocalMovieGenreSchema:
	// 	return ProcessLocalMovieGenre(content, dbConnection, apiClient)
	// case SharedModels.LocalSectionSchema:
	// 	return ProcessLocalSection(content, dbConnection)
	// case SharedModels.LocalPollSchema:
	// 	return ProcessLocalPoll(content, dbConnection)
	case SharedModels.LocalMovieSchema:
		return ProcessLocalMovie(content, dbConnection, apiClient)
	default:
		log.Printf("Cannot perform specific action for type %T", v)
	}

	return nil
}

func ProcessLocalMovie(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiClient *ApiClient.APIClient) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
	defer cancel()

	localMovie := SharedModels.Movie{}
	detail := content.Details.(SharedModels.LocalMovieSchema)
	localMovie.ContentId = content.ID
	if content.Enable {

		movieDetail, err := apiClient.GetMovieDetail(int(detail.MovieID))
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DOWNLOAD_ERROR, err)
		}

		localMovie.Ages = &movieDetail.Ages
		localMovie.Casts = movieDetail.Casts
		localMovie.Company = &movieDetail.Company
		localMovie.Description = movieDetail.Description
		localMovie.Duration = &movieDetail.Duration
		localMovie.EntityId = &detail.MovieID

		localMovie.Genres = movieDetail.Genres
		localMovie.ImdbCode = &movieDetail.IMDBCode
		localMovie.ImdbRate = movieDetail.IMDBRate
		//TODO: download the video if the extracted content does not exist on fs
		extractedPath, podspaceHash, err := DownloadZippedVideo(apiClient, detail.FileLink, "")
		if err != nil {
			return err
		}

		entries, err := os.ReadDir(extractedPath)
		if err != nil {
			return cstmerr.NewProcessError(fmt.Sprintf(cstmerr.PROCESS_FIND_DIRECTORY, extractedPath), err)
		}

		var destinationFile string
		var destinationSub string
		for _, entry := range entries {
			if entry.IsDir() {
				destinationSub = entry.Name()
			}
		}

		if len(destinationSub) == 0 {
			return cstmerr.NewProcessError(cstmerr.PROCESS_CREATE_ERROR, nil)
		}

		masterFile := fmt.Sprintf("%s/master_%s.m3u8", destinationSub, destinationSub)
		destinationFile = filepath.Join(extractedPath, masterFile)

		hash, err := SharedModels.CalculateMD5(destinationFile, 1025)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_HASH_ERROR, err)
		}
		localMovie.Link.FileHash = hex.EncodeToString(hash)
		localMovie.Link.PlayLink = filepath.Join(podspaceHash[0:len(podspaceHash)-4], masterFile)
		log.Printf("debug: playlink %s", localMovie.Link.PlayLink)

		localMovie.NameEn = &movieDetail.NameEn
		localMovie.NameFa = movieDetail.NameFa

		localMovie.PostId = movieDetail.PostID
		localMovie.YearsOfBroadcast = &movieDetail.YearsOFBroadcast

		_, bannerUrlPodspaceHash, err := DownloadImage(apiClient, movieDetail.BannerURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, movieDetail.BannerURL), err)
		}
		localMovie.Image.BannerUrl = &bannerUrlPodspaceHash

		_, imageUrlPodspaceHash, err := DownloadImage(apiClient, movieDetail.ImageURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, movieDetail.ImageURL), err)
		}
		localMovie.Image.ImageURL = imageUrlPodspaceHash

		_, mobileBannerUrlPodspaceHash, err := DownloadImage(apiClient, movieDetail.MobileBannerURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, movieDetail.MobileBannerURL), err)
		}
		localMovie.Image.MobileBannerUrl = &mobileBannerUrlPodspaceHash

		err = dbConnection.Save(ctx, &localMovie)
		if err != nil {
			return cstmerr.NewProcessError("failed to create slider", err)
		}

	} else {

	}
	return nil
}

func ProcessLocalPoll(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
	defer cancel()
	localPoll := SharedModels.Poll{}
	detail := content.Details.(SharedModels.LocalPollSchema)
	localPoll.ContentId = content.ID
	if content.Enable {
		localPoll.Questions = detail.Questions
		//TODO: add title

		err := dbConnection.Save(ctx, &localPoll)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_CREATE_ERROR, err)
		}
	} else {
		err := dbConnection.Delete(ctx, &localPoll)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DELETE_ENTITY, err)
		}
	}

	return nil
}

func ProcessLocalSection(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
	defer cancel()

	localSection := SharedModels.Section{}
	detail := content.Details.(SharedModels.LocalSectionSchema)
	localSection.ContentId = content.ID

	if content.Enable {
		localSection.Action = detail.Action
		localSection.CardType = detail.CardType
		localSection.Name = &detail.Name
		priority := int32(detail.Priority)
		localSection.Priority = &priority
		//TODO: what is entity type

		err := dbConnection.Save(ctx, &localSection)
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

			err = dbConnection.CreateAssosiate(ctx, &localSection, "Tabs", &tabs)
			if err != nil {
				return cstmerr.NewProcessError("failed to create assosiate tab page", err)
			}
		}
	} else {

	}

	return nil
}

func ProcessLocalMovieGenre(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiclient *ApiClient.APIClient) error {

	const GENRE = "genre"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	localMovieGenre := SharedModels.Genre{}
	detail := content.Details.(SharedModels.LocalMovieGenreSchema)
	localMovieGenre.ContentId = content.ID
	if content.Enable {

		localMovieGenre.Code = detail.Code
		localMovieGenre.Enable = content.Enable
		//TODO: get name

		_, imageUrlPodspaceHash, err := DownloadImage(apiclient, detail.ImageURL, GENRE)
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, detail.ImageURL), err)
		}
		trick := filepath.Join(GENRE, imageUrlPodspaceHash)
		localMovieGenre.ImageURL = &trick

		err = dbConnection.Save(ctx, &localMovieGenre)
		if err != nil {
			return cstmerr.NewProcessError("failed to create slider", err)
		}
	} else {
		//TODO: handle image deletion from filespace
		err := dbConnection.Delete(ctx, &localMovieGenre)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DELETE_ENTITY, err)
		}
	}

	return nil
}

func ProcessLocalSlider(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiclient *ApiClient.APIClient) error {
	const SLIDER = "slider"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
	defer cancel()

	localSlider := SharedModels.Slider{}
	detail := content.Details.(SharedModels.LocalSliderSchema)
	localSlider.ContentId = content.ID

	if content.Enable {

		localSlider.ButtonTitle = detail.ButtonTitle

		_, imageUrlPodspaceHash, err := DownloadImage(apiclient, detail.ImageURL, SLIDER)
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, detail.ImageURL), err)
		}
		localSlider.Image.ImageURL = filepath.Join(SLIDER, imageUrlPodspaceHash)

		if detail.LogoImageURL != nil {
			_, logoImageUrlPodspaceHash, err := DownloadImage(apiclient, *detail.LogoImageURL, SLIDER)
			if err != nil {
				return cstmerr.NewProcessError(
					fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, detail.ImageURL), err)
			}
			trick := filepath.Join(SLIDER, logoImageUrlPodspaceHash)
			localSlider.Image.LogoImageUrl = &trick
		}

		_, mediumImageUrlPodspaceHash, err := DownloadImage(apiclient, detail.MediumImageURL, SLIDER)
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, detail.ImageURL), err)
		}
		trick := filepath.Join(SLIDER, mediumImageUrlPodspaceHash)
		localSlider.Image.MediumImageUrl = &trick

		_, smallImageUrlPodspaceHash, err := DownloadImage(apiclient, detail.SmallImageURL, SLIDER)
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, detail.ImageURL), err)
		}
		trick2 := filepath.Join(SLIDER, smallImageUrlPodspaceHash)
		localSlider.Image.SmallImageUrl = &trick2

		localSlider.Link = detail.Link

		err = dbConnection.Save(ctx, &localSlider)
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

			err = dbConnection.CreateAssosiate(ctx, &localSlider, "Tabs", &tabs)
			if err != nil {
				return cstmerr.NewProcessError("failed to create assosiate tab page", err)
			}
		}
	} else {
		//TODO: handle assosiation
		err := dbConnection.Delete(ctx, &localSlider)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DELETE_ENTITY, err)
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

	if content.Enable {

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
			err := dbConnection.CreateAssosiate(ctx, &localTab, "Pages", &pages)
			if err != nil {
				return cstmerr.NewProcessError("failed to create assosiate tab page", err)
			}
		}
	} else {
		//TODO: handle assosiation
		err := dbConnection.Delete(ctx, &localTab)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DELETE_ENTITY, err)
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
	if content.Enable {
		localPage.Name = &detail.Name
		localPage.Type = detail.Type
		err := dbConnection.Save(ctx, &localPage)
		if err != nil {
			return cstmerr.NewProcessError("failed to save Local Page", err)
		}
	} else {
		err := dbConnection.Delete(ctx, &localPage)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DELETE_ENTITY, err)
		}
	}
	return nil
}

func ProcessLocalAdvertisement(
	content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiclient *ApiClient.APIClient) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
	defer cancel()
	localAdvertisement := SharedModels.Advertisement{}
	localAdvertisementLink := SharedModels.AdvertisementLink{}
	localAdvertisement.ContentId = content.ID
	if content.Enable {
		detail := content.Details.(SharedModels.LocalAdvertisementSchema)
		// Download filelink to destination
		destinationFile, podspaceHash, err := DownloadVideo(apiclient, detail.FileLink, "ads")
		if err != nil {
			return err
		}
		localAdvertisement.SkipDuration = int32(detail.SkipDuration)
		localAdvertisement.Synced = false
		localAdvertisementLink.LinkType = "MP4"
		hash, err := SharedModels.CalculateMD5(destinationFile, 1025)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_HASH_ERROR, err)
		}
		localAdvertisementLink.FileHash = hex.EncodeToString(hash)
		localAdvertisementLink.PlayLink = filepath.Join("ads", podspaceHash)
		localAdvertisementLink.OriginalLink = detail.FileLink
		localAdvertisement.Link = localAdvertisementLink
		dbConnection.Save(ctx, &localAdvertisement)
	} else {
		//TODO: handle file deletion from filespace
		err := dbConnection.First(ctx, &localAdvertisement)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DELETE_ENTITY, err)
		}

		err = DeleteVideo(localAdvertisement.Link.PlayLink)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DELETE_FILE, err)
		}

		err = dbConnection.Delete(ctx, &localAdvertisement)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DELETE_ENTITY, err)
		}

	}
	return nil
}
