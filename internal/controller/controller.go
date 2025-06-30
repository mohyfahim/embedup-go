package controller

import (
	"context"
	ApiClient "embedup-go/internal/apiclient"
	"embedup-go/internal/cstmerr"
	"embedup-go/internal/dbclient"
	SharedModels "embedup-go/internal/shared"
	"encoding/hex"
	"fmt"
	"io/fs"
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

	fileInformation := SharedModels.FileInformation{}
	err = apiclient.GetFileInformation(url, &fileInformation)

	if err != nil {
		return "", "", cstmerr.NewProcessError(cstmerr.PROCESS_FILE_INFO, err)
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

	fileInformation := SharedModels.FileInformation{}
	err = apiclient.GetFileInformation(url, &fileInformation)

	if err != nil {
		return "", "", cstmerr.NewProcessError(cstmerr.PROCESS_FILE_INFO, err)
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

func DownloadAudio(apiclient *ApiClient.APIClient, url string, dir ...string) (string, string, error) {

	contentBasePath := os.Getenv("PODBOX_UPDATE_CONTENT_BASE_PATH")
	if contentBasePath == "" {
		contentBasePath = "/mnt/sdcard/assets/"
	}
	destinationPath := filepath.Join(append([]string{contentBasePath, "audios"}, dir...)...)

	log.Printf("destination path for download file : %s \n", destinationPath)
	err := SharedModels.CheckAndCreateDir(destinationPath)
	if err != nil {
		log.Printf("Error in creating path %s: %v", destinationPath, err)
	}

	fileInformation := SharedModels.FileInformation{}
	err = apiclient.GetFileInformation(url, &fileInformation)

	if err != nil {
		return "", "", cstmerr.NewProcessError(cstmerr.PROCESS_FILE_INFO, err)
	}

	fileNameWithPrefix := fileInformation.MD5 + ".mp3"

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

func GetDirectorySize(path string) (int64, error) {
	var totalSize int64
	err := filepath.Walk(path, func(filePath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("error walking directory: %w", err)
	}
	return totalSize, nil
}

func CheckExtractedExist(apiclient *ApiClient.APIClient,
	downloadLink string, fileInformation *SharedModels.FileInformation) (bool, error) {
	contentBasePath := os.Getenv("PODBOX_UPDATE_CONTENT_BASE_PATH")
	if contentBasePath == "" {
		contentBasePath = "/mnt/sdcard/assets/"
	}
	destinationPath := filepath.Join(contentBasePath, "videos")

	err := apiclient.GetFileInformation(downloadLink, fileInformation)

	if err != nil {
		return false, cstmerr.NewProcessError(cstmerr.PROCESS_FILE_INFO, err)
	}

	fileInformation.FileNameWithPrefix = fileInformation.MD5 + ".zip"
	fileInformation.DestinationPath = filepath.Join(destinationPath, fileInformation.FileNameWithPrefix)
	fileInformation.DestinationExtracted = filepath.Join(destinationPath, fileInformation.MD5)

	info, err := os.Stat(fileInformation.DestinationExtracted)
	if err != nil && os.IsNotExist(err) {
		log.Printf("Error in finding directory %s: %v\n", fileInformation.DestinationExtracted, err)
		return false, nil
	} else if err != nil {
		log.Printf("Error in finding directory %s: %v\n", fileInformation.DestinationExtracted, err)
		return false, err
	} else if info.IsDir() {
		log.Printf("Directory '%s' exists.\n", fileInformation.DestinationExtracted)
		return true, nil
	}
	log.Printf("Error in finding directory %s: %v\n", fileInformation.DestinationExtracted, err)
	return false, nil
}

func DownloadZippedVideo(apiclient *ApiClient.APIClient, url string,
	fileInformation *SharedModels.FileInformation, dir ...string) error {

	err := apiclient.DownloadFileWithRetry(url, fileInformation.DestinationPath)

	if err != nil {
		log.Printf("error in downloading hash")
		return cstmerr.NewDownloadError(
			fmt.Sprintf("failed to download multiple times: %s", url))
	}

	err = SharedModels.UnzipFile(fileInformation.DestinationPath, fileInformation.DestinationExtracted)
	if err != nil {
		return err
	}
	return nil
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
	// case SharedModels.LocalAdvertisementSchema:
	// 	return ProcessLocalAdvertisement(content, dbConnection, apiClient)
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
	// case SharedModels.LocalMovieSchema:
	// 	return ProcessLocalMovie(content, dbConnection, apiClient)
	// case SharedModels.LocalSeriesSchema:
	// 	return ProcessLocalSeries(content, dbConnection, apiClient)
	// case SharedModels.LocalSeriesSeasonSchema:
	// 	return ProcessLocalSeriesSeason(content, dbConnection, apiClient)
	// case SharedModels.LocalSeriesEpisodeSchema:
	// 	return ProcessLocalEpisodeSeason(content, dbConnection, apiClient)
	// case SharedModels.LocalSectionContentSchema:
	// 	return ProcessLocalSectionContent(content, dbConnection)
	// case SharedModels.LocalPodcastParentSchema:
	// 	return ProcessLocalPodcastParent(content, dbConnection, apiClient)
	// case SharedModels.LocalAudiobookParentSchema:
	// 	return ProcessLocalAudiobookParent(content, dbConnection, apiClient)
	// case SharedModels.LocalPodcastSchema:
	// 	return ProcessLocalPodcast(content, dbConnection, apiClient)
	// case SharedModels.LocalAudiobookSchema:
	// 	return ProcessLocalAudiobook(content, dbConnection, apiClient)
	// case SharedModels.LocalAlbumSchema:
	// 	return ProcessLocalAlbum(content, dbConnection, apiClient)
	// case SharedModels.LocalMusicSchema:
	// 	return ProcessLocalMusic(content, dbConnection, apiClient)
	case SharedModels.LocalTermsConditionsSchema:
		return ProcessLocalTerms(content, dbConnection)
	case SharedModels.LocalMagazineSchema:
		return ProcessLocalMagazine(content, dbConnection, apiClient)
	case SharedModels.LocalNewsSchema:
		return ProcessLocalNews(content, dbConnection, apiClient)
	default:
		log.Printf("Cannot perform specific action for type %T", v)
	}

	return nil
}

func ProcessLocalNews(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiClient *ApiClient.APIClient) error {
	localNews := SharedModels.News{}
	detail := content.Details.(SharedModels.LocalNewsSchema)
	localNews.ContentId = content.ID
	if content.Enable {

		if detail.ImageURL != nil {
			_, imageUrlPodspaceHash, err := DownloadImage(apiClient, *detail.ImageURL, "")
			if err != nil {
				return cstmerr.NewProcessError(
					fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, *detail.ImageURL), err)
			}
			localNews.Image.ImageURL = &imageUrlPodspaceHash
		}
		if detail.BannerURL != nil {
			_, bannerUrlPodspaceHash, err := DownloadImage(apiClient, *detail.BannerURL, "")
			if err != nil {
				return cstmerr.NewProcessError(
					fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, *detail.BannerURL), err)
			}
			localNews.Image.BannerUrl = &bannerUrlPodspaceHash
		}

		localNews.LongText = detail.LongText
		localNews.Text = &detail.Text

		publishDate := time.UnixMilli(int64(detail.ReleaseTime))
		localNews.ReleaseTime = &publishDate

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
		defer cancel()

		err := dbConnection.Save(ctx, &localNews)
		if err != nil {
			return cstmerr.NewProcessError("failed to create news", err)
		}

	} else {
		panic("unimplemented")
	}
	return nil
}

func ProcessLocalMagazine(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiClient *ApiClient.APIClient) error {
	localMag := SharedModels.Magazine{}
	detail := content.Details.(SharedModels.LocalMagazineSchema)
	localMag.ContentId = content.ID
	if content.Enable {

		if detail.ImageURL != nil {
			_, imageUrlPodspaceHash, err := DownloadImage(apiClient, *detail.ImageURL, "")
			if err != nil {
				return cstmerr.NewProcessError(
					fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, *detail.ImageURL), err)
			}
			localMag.Image.ImageURL = &imageUrlPodspaceHash
		}
		if detail.BannerURL != nil {
			_, bannerUrlPodspaceHash, err := DownloadImage(apiClient, *detail.BannerURL, "")
			if err != nil {
				return cstmerr.NewProcessError(
					fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, *detail.BannerURL), err)
			}
			localMag.Image.BannerUrl = &bannerUrlPodspaceHash
		}

		localMag.LongText = detail.LongText
		localMag.Text = &detail.Text

		publishDate := time.UnixMilli(int64(detail.ReleaseTime))
		localMag.ReleaseTime = &publishDate

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
		defer cancel()

		err := dbConnection.Save(ctx, &localMag)
		if err != nil {
			return cstmerr.NewProcessError("failed to create news", err)
		}

	} else {
		panic("unimplemented")
	}
	return nil
}

func ProcessLocalTerms(content SharedModels.ProcessedContentSchema, dbConnection dbclient.DBClient) error {
	localTerm := SharedModels.TermsConditions{}
	detail := content.Details.(SharedModels.LocalTermsConditionsSchema)
	localTerm.ContentId = content.ID
	if content.Enable {
		localTerm.Content = detail.Content
		localTerm.Name = detail.Name

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
		defer cancel()

		err := dbConnection.Save(ctx, &localTerm)
		if err != nil {
			return cstmerr.NewProcessError("failed to create audiobook parent", err)
		}

	} else {
		panic("unimplemented")
	}
	return nil
}

func ProcessLocalMusic(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiClient *ApiClient.APIClient) error {
	localMusic := SharedModels.Music{}
	detail := content.Details.(SharedModels.LocalMusicSchema)
	localMusic.ContentId = content.ID
	if content.Enable {
		musicDetail, err := apiClient.GetMusicDetail(int(detail.MusicID))
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DOWNLOAD_ERROR, err)
		}
		localMusic.EntityId = musicDetail.ID
		localMusic.Description = musicDetail.Description
		localMusic.AlbumContentId = detail.LocalAlbumID
		localMusic.Genres = musicDetail.Genre
		localMusic.Name = musicDetail.NameFa
		if musicDetail.PublishDate != nil {
			publishDate := time.UnixMilli(int64(*musicDetail.PublishDate))
			localMusic.PublishDate = &publishDate
		}

		for _, agent := range musicDetail.Agents {
			_, bannerUrlAgentPodspaceHash, err := DownloadImage(apiClient, agent.BannerUrl, "")
			if err != nil {
				return cstmerr.NewProcessError(
					fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, musicDetail.ImageURL), err)
			}
			agent.BannerUrl = bannerUrlAgentPodspaceHash
		}
		localMusic.Agents = musicDetail.Agents

		_, imageUrlPodspaceHash, err := DownloadImage(apiClient, musicDetail.ImageURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, musicDetail.ImageURL), err)
		}
		localMusic.Image.ImageURL = &imageUrlPodspaceHash

		_, bannerUrlPodspaceHash, err := DownloadImage(apiClient, musicDetail.BannerURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, musicDetail.BannerURL), err)
		}
		localMusic.Image.BannerUrl = &bannerUrlPodspaceHash

		destinationFile, podspaceHash, err := DownloadAudio(apiClient, detail.FileLink)
		if err != nil {
			return err
		}

		hash, err := SharedModels.CalculateMD5(destinationFile, 1025)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_HASH_ERROR, err)
		}
		localMusic.Link.FileHash = hex.EncodeToString(hash)
		localMusic.Link.PlayLink = podspaceHash

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
		defer cancel()

		err = dbConnection.Save(ctx, &localMusic)
		if err != nil {
			return cstmerr.NewProcessError("failed to create audiobook parent", err)
		}

	} else {
		panic("unimplemented")
	}
	return nil
}

func ProcessLocalAlbum(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiClient *ApiClient.APIClient) error {
	localAlbum := SharedModels.Album{}
	detail := content.Details.(SharedModels.LocalAlbumSchema)
	localAlbum.ContentId = content.ID
	if content.Enable {
		albumDetail, err := apiClient.GetAlbumDetail(int(detail.AlbumID))
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DOWNLOAD_ERROR, err)
		}
		localAlbum.EntityId = albumDetail.ID
		localAlbum.Description = albumDetail.Description
		localAlbum.Name = albumDetail.NameFa
		localAlbum.Genre = albumDetail.Genre

		for _, agent := range albumDetail.Agents {
			_, bannerUrlAgentPodspaceHash, err := DownloadImage(apiClient, agent.BannerUrl, "")
			if err != nil {
				return cstmerr.NewProcessError(
					fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, albumDetail.ImageURL), err)
			}
			agent.BannerUrl = bannerUrlAgentPodspaceHash
		}
		localAlbum.Agents = albumDetail.Agents

		_, imageUrlPodspaceHash, err := DownloadImage(apiClient, albumDetail.ImageURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, albumDetail.ImageURL), err)
		}
		localAlbum.Image.ImageURL = &imageUrlPodspaceHash

		_, bannerUrlPodspaceHash, err := DownloadImage(apiClient, albumDetail.BannerURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, albumDetail.BannerURL), err)
		}
		localAlbum.Image.BannerUrl = &bannerUrlPodspaceHash

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
		defer cancel()

		err = dbConnection.Save(ctx, &localAlbum)
		if err != nil {
			return cstmerr.NewProcessError("failed to create audiobook parent", err)
		}

	} else {

		panic("not implemented")
	}
	return nil
}
func ProcessLocalAudiobook(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiClient *ApiClient.APIClient) error {

	localAudiobook := SharedModels.AudioBook{}
	detail := content.Details.(SharedModels.LocalAudiobookSchema)
	localAudiobook.ContentId = content.ID
	if content.Enable {
		audiobookDetail, err := apiClient.GetAudiobookDetail(int(detail.AudiobookID))
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DOWNLOAD_ERROR, err)
		}

		localAudiobook.EntityId = audiobookDetail.ID
		localAudiobook.Ages = audiobookDetail.Ages
		localAudiobook.Duration = audiobookDetail.Duration
		localAudiobook.Name = audiobookDetail.NameFa
		localAudiobook.Genre = audiobookDetail.Genre

		for _, agent := range audiobookDetail.Agents {
			_, bannerUrlAgentPodspaceHash, err := DownloadImage(apiClient, agent.BannerUrl, "")
			if err != nil {
				return cstmerr.NewProcessError(
					fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, audiobookDetail.ImageURL), err)
			}
			agent.BannerUrl = bannerUrlAgentPodspaceHash
		}
		localAudiobook.Agents = audiobookDetail.Agents

		publishDate := time.UnixMilli(int64(audiobookDetail.PublishDate))
		localAudiobook.PublishDate = &publishDate
		localAudiobook.AudiobookAlbumContentId = &detail.LocalAudiobookParentID

		_, imageUrlPodspaceHash, err := DownloadImage(apiClient, audiobookDetail.ImageURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, audiobookDetail.ImageURL), err)
		}
		localAudiobook.Image.ImageURL = &imageUrlPodspaceHash

		_, bannerUrlPodspaceHash, err := DownloadImage(apiClient, audiobookDetail.BannerURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, audiobookDetail.BannerURL), err)
		}
		localAudiobook.Image.BannerUrl = &bannerUrlPodspaceHash

		destinationFile, podspaceHash, err := DownloadAudio(apiClient, detail.FileLink)
		if err != nil {
			return err
		}

		hash, err := SharedModels.CalculateMD5(destinationFile, 1025)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_HASH_ERROR, err)
		}
		localAudiobook.Link.FileHash = hex.EncodeToString(hash)
		localAudiobook.Link.PlayLink = podspaceHash

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
		defer cancel()

		err = dbConnection.Save(ctx, &localAudiobook)
		if err != nil {
			return cstmerr.NewProcessError("failed to create audiobook parent", err)
		}

	} else {
		panic("unimplemented")
	}
	return nil
}

func ProcessLocalPodcast(content SharedModels.ProcessedContentSchema, dbConnection dbclient.DBClient, apiClient *ApiClient.APIClient) error {

	localPodcast := SharedModels.Podcast{}
	detail := content.Details.(SharedModels.LocalPodcastSchema)
	localPodcast.ContentId = content.ID
	if content.Enable {
		podcastDetail, err := apiClient.GetPodcastDetail(int(detail.PodcastID))
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DOWNLOAD_ERROR, err)
		}

		localPodcast.EntityId = podcastDetail.ID
		localPodcast.Description = podcastDetail.Description
		localPodcast.Ages = podcastDetail.Ages
		localPodcast.Duration = podcastDetail.Duration
		localPodcast.Name = podcastDetail.NameFa
		localPodcast.Genre = podcastDetail.Genre

		for _, agent := range podcastDetail.Agents {
			_, bannerUrlAgentPodspaceHash, err := DownloadImage(apiClient, agent.BannerUrl, "")
			if err != nil {
				return cstmerr.NewProcessError(
					fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, podcastDetail.ImageURL), err)
			}
			agent.BannerUrl = bannerUrlAgentPodspaceHash
		}
		localPodcast.Agents = podcastDetail.Agents

		publishDate := time.UnixMilli(int64(podcastDetail.PublishDate))
		localPodcast.PublishDate = &publishDate
		localPodcast.PodcastAlbumContentId = &detail.LocalPodcastParentID

		_, imageUrlPodspaceHash, err := DownloadImage(apiClient, podcastDetail.ImageURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, podcastDetail.ImageURL), err)
		}
		localPodcast.Image.ImageURL = &imageUrlPodspaceHash

		_, bannerUrlPodspaceHash, err := DownloadImage(apiClient, podcastDetail.BannerURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, podcastDetail.BannerURL), err)
		}
		localPodcast.Image.BannerUrl = &bannerUrlPodspaceHash

		destinationFile, podspaceHash, err := DownloadAudio(apiClient, detail.FileLink)
		if err != nil {
			return err
		}

		hash, err := SharedModels.CalculateMD5(destinationFile, 1025)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_HASH_ERROR, err)
		}
		localPodcast.Link.FileHash = hex.EncodeToString(hash)
		localPodcast.Link.PlayLink = podspaceHash

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
		defer cancel()

		err = dbConnection.Save(ctx, &localPodcast)
		if err != nil {
			return cstmerr.NewProcessError("failed to create audiobook parent", err)
		}

	} else {
		panic("unimplemented")
	}
	return nil
}

func ProcessLocalAudiobookParent(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiClient *ApiClient.APIClient) error {

	localAudiobookParent := SharedModels.AudiobookAlbum{}
	detail := content.Details.(SharedModels.LocalAudiobookParentSchema)
	localAudiobookParent.ContentId = content.ID
	if content.Enable {
		audiobookParentDetail, err := apiClient.GetAudiobookParentDetail(int(detail.AudiobookParentID))
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DOWNLOAD_ERROR, err)
		}

		localAudiobookParent.Description = audiobookParentDetail.Description
		publishDate := time.UnixMilli(int64(audiobookParentDetail.PublishDate))
		localAudiobookParent.PublishDate = &publishDate
		localAudiobookParent.EntityId = audiobookParentDetail.ID
		localAudiobookParent.Duration = audiobookParentDetail.Duration
		localAudiobookParent.Name = audiobookParentDetail.NameFa
		localAudiobookParent.Genre = audiobookParentDetail.Genre

		for _, agent := range audiobookParentDetail.Agents {
			_, bannerUrlAgentPodspaceHash, err := DownloadImage(apiClient, agent.BannerUrl, "")
			if err != nil {
				return cstmerr.NewProcessError(
					fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, audiobookParentDetail.ImageURL), err)
			}
			agent.BannerUrl = bannerUrlAgentPodspaceHash
		}
		localAudiobookParent.Agents = audiobookParentDetail.Agents

		_, imageUrlPodspaceHash, err := DownloadImage(apiClient, audiobookParentDetail.ImageURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, audiobookParentDetail.ImageURL), err)
		}
		localAudiobookParent.Image.ImageURL = &imageUrlPodspaceHash

		_, bannerUrlPodspaceHash, err := DownloadImage(apiClient, audiobookParentDetail.BannerURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, audiobookParentDetail.ImageURL), err)
		}
		localAudiobookParent.Image.BannerUrl = &bannerUrlPodspaceHash

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
		defer cancel()

		err = dbConnection.Save(ctx, &localAudiobookParent)
		if err != nil {
			return cstmerr.NewProcessError("failed to create audiobook parent", err)
		}

	} else {
		panic("unimplemented")

	}

	return nil
}

func ProcessLocalPodcastParent(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiClient *ApiClient.APIClient) error {

	localPodcastParent := SharedModels.PodcastAlbum{}
	detail := content.Details.(SharedModels.LocalPodcastParentSchema)
	localPodcastParent.ContentId = content.ID
	if content.Enable {
		podcastParentDetail, err := apiClient.GetPodcastParentDetail(int(detail.PodcastParentID))
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DOWNLOAD_ERROR, err)
		}
		localPodcastParent.EntityId = podcastParentDetail.ID
		publishDate := time.UnixMilli(int64(podcastParentDetail.PublishDate))
		localPodcastParent.PublishDate = &publishDate
		localPodcastParent.Duration = podcastParentDetail.Duration
		localPodcastParent.Name = podcastParentDetail.NameFa
		localPodcastParent.Genre = podcastParentDetail.Genre
		//TODO: loop over agents and change the image url to image path
		for _, agent := range podcastParentDetail.Agents {
			_, bannerUrlAgentPodspaceHash, err := DownloadImage(apiClient, agent.BannerUrl, "")
			if err != nil {
				return cstmerr.NewProcessError(
					fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, podcastParentDetail.ImageURL), err)
			}
			agent.BannerUrl = bannerUrlAgentPodspaceHash
		}

		localPodcastParent.Agents = podcastParentDetail.Agents

		_, imageUrlPodspaceHash, err := DownloadImage(apiClient, podcastParentDetail.ImageURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, podcastParentDetail.ImageURL), err)
		}
		localPodcastParent.Image.ImageURL = &imageUrlPodspaceHash

		_, bannerUrlPodspaceHash, err := DownloadImage(apiClient, podcastParentDetail.BannerURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, podcastParentDetail.ImageURL), err)
		}
		localPodcastParent.Image.BannerUrl = &bannerUrlPodspaceHash

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
		defer cancel()

		err = dbConnection.Save(ctx, &localPodcastParent)
		if err != nil {
			return cstmerr.NewProcessError("failed to create podcast parent", err)
		}

	} else {
		panic("unimplemented")
	}
	return nil
}

func ProcessLocalSectionContent(content SharedModels.ProcessedContentSchema, dbConnection dbclient.DBClient) error {

	localSectionContent := SharedModels.SectionContent{}
	detail := content.Details.(SharedModels.LocalSectionContentSchema)
	localSectionContent.ContentId = content.ID
	if content.Enable {
		localSectionContent.EntityContentId = detail.EntityContentID
		localSectionContent.EntityContentType = detail.EntityContentType
		localSectionContent.Priority = detail.Priority
		localSectionContent.SectionContentId = &detail.LocalSectionID

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
		defer cancel()

		err := dbConnection.Save(ctx, &localSectionContent)
		if err != nil {
			return cstmerr.NewProcessError("failed to create section content", err)
		}

	} else {
		panic("not implemented")
	}
	return nil
}

func ProcessLocalEpisodeSeason(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiClient *ApiClient.APIClient) error {

	localSeriesEpisode := SharedModels.SeriesEpisode{}
	detail := content.Details.(SharedModels.LocalSeriesEpisodeSchema)
	localSeriesEpisode.ContentId = content.ID
	if content.Enable {
		seriesEpisodeDetail, err := apiClient.GetSeriesEpisodeDetail(int(detail.EpisodeID))
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DOWNLOAD_ERROR, err)
		}
		localSeriesEpisode.EntityId = &seriesEpisodeDetail.ID
		localSeriesEpisode.Index = int64(seriesEpisodeDetail.Index)
		localSeriesEpisode.Name = seriesEpisodeDetail.Name
		localSeriesEpisode.NameEn = &seriesEpisodeDetail.NameEn
		localSeriesEpisode.SeasonContentId = &detail.LocalSeasonID

		fileInformation := SharedModels.FileInformation{}

		checkExtracted, err := CheckExtractedExist(apiClient, detail.FileLink, &fileInformation)
		if err != nil {
			return err
		}
		if !checkExtracted {
			err := DownloadZippedVideo(apiClient, detail.FileLink, &fileInformation, "")
			if err != nil {
				return err
			}
		}

		entries, err := os.ReadDir(fileInformation.DestinationExtracted)
		if err != nil {
			return cstmerr.NewProcessError(fmt.Sprintf(cstmerr.PROCESS_FIND_DIRECTORY,
				fileInformation.DestinationExtracted), err)
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
		destinationFile = filepath.Join(fileInformation.DestinationExtracted, masterFile)

		hash, err := SharedModels.CalculateMD5(destinationFile, 1025)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_HASH_ERROR, err)
		}
		localSeriesEpisode.Link.FileHash = hex.EncodeToString(hash)
		localSeriesEpisode.Link.PlayLink = filepath.Join(
			fileInformation.FileNameWithPrefix[0:len(fileInformation.FileNameWithPrefix)-4], masterFile)
		log.Printf("debug: playlink %s", localSeriesEpisode.Link.PlayLink)

		_, imageUrlPodspaceHash, err := DownloadImage(apiClient, seriesEpisodeDetail.ImageURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, seriesEpisodeDetail.ImageURL), err)
		}
		localSeriesEpisode.Image.ImageURL = &imageUrlPodspaceHash

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
		defer cancel()

		err = dbConnection.Save(ctx, &localSeriesEpisode)
		if err != nil {
			return cstmerr.NewProcessError("failed to create season", err)
		}

	} else {
		panic("not implemented")
	}
	return nil
}

func ProcessLocalSeriesSeason(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiClient *ApiClient.APIClient) error {
	localSeriesSeason := SharedModels.SeriesSeason{}
	detail := content.Details.(SharedModels.LocalSeriesSeasonSchema)
	localSeriesSeason.ContentId = content.ID
	if content.Enable {
		seriesSeasonDetail, err := apiClient.GetSeriesSeasonDetail(int(detail.SeasonID))
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DOWNLOAD_ERROR, err)
		}
		localSeriesSeason.Index = int64(seriesSeasonDetail.Index)
		localSeriesSeason.Name = seriesSeasonDetail.Name
		localSeriesSeason.NameEn = &seriesSeasonDetail.NameEn
		localSeriesSeason.EntityId = &seriesSeasonDetail.ID

		localSeriesSeason.SeriesContentId = &detail.LocalSeriesID

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
		defer cancel()

		err = dbConnection.Save(ctx, &localSeriesSeason)
		if err != nil {
			return cstmerr.NewProcessError("failed to create season", err)
		}

	} else {
		panic("not implemented")

	}
	return nil
}

func ProcessLocalSeries(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiClient *ApiClient.APIClient) error {
	localSeries := SharedModels.Series{}
	detail := content.Details.(SharedModels.LocalSeriesSchema)
	localSeries.ContentId = content.ID
	if content.Enable {
		seriesDetail, err := apiClient.GetSeriesDetail(int(detail.SeriesID))
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_DOWNLOAD_ERROR, err)
		}
		localSeries.Ages = &seriesDetail.Ages
		localSeries.Casts = seriesDetail.Casts
		localSeries.Company = &seriesDetail.Company
		localSeries.Description = seriesDetail.Description
		localSeries.EntityId = &seriesDetail.ID

		localSeries.Genres = seriesDetail.Genres
		localSeries.ImdbCode = &seriesDetail.IMDBCode
		localSeries.ImdbRate = seriesDetail.IMDBRate
		localSeries.NameEn = &seriesDetail.NameEn
		localSeries.NameFa = seriesDetail.NameFa
		localSeries.PostId = seriesDetail.PostID
		localSeries.YearsOfBroadcast = &seriesDetail.YearsOFBroadcast

		_, bannerUrlPodspaceHash, err := DownloadImage(apiClient, seriesDetail.BannerURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, seriesDetail.BannerURL), err)
		}
		localSeries.Image.BannerUrl = &bannerUrlPodspaceHash

		_, imageUrlPodspaceHash, err := DownloadImage(apiClient, seriesDetail.ImageURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, seriesDetail.ImageURL), err)
		}
		localSeries.Image.ImageURL = imageUrlPodspaceHash

		_, mobileBannerUrlPodspaceHash, err := DownloadImage(apiClient, seriesDetail.MobileBannerURL, "")
		if err != nil {
			return cstmerr.NewProcessError(
				fmt.Sprintf(cstmerr.PROCESS_DOWNLOAD_ERROR, seriesDetail.MobileBannerURL), err)
		}
		localSeries.Image.MobileBannerUrl = &mobileBannerUrlPodspaceHash

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
		defer cancel()

		err = dbConnection.Save(ctx, &localSeries)
		if err != nil {
			return cstmerr.NewProcessError("failed to create slider", err)
		}

	} else {
		panic("not implemented")

	}
	return nil
}

func ProcessLocalMovie(content SharedModels.ProcessedContentSchema,
	dbConnection dbclient.DBClient, apiClient *ApiClient.APIClient) error {

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

		fileInformation := SharedModels.FileInformation{}

		checkExtracted, err := CheckExtractedExist(apiClient, detail.FileLink, &fileInformation)
		if err != nil {
			return err
		}
		if !checkExtracted {
			err := DownloadZippedVideo(apiClient, detail.FileLink, &fileInformation, "")
			if err != nil {
				return err
			}
		}

		entries, err := os.ReadDir(fileInformation.DestinationExtracted)
		if err != nil {
			return cstmerr.NewProcessError(fmt.Sprintf(cstmerr.PROCESS_FIND_DIRECTORY,
				fileInformation.DestinationExtracted), err)
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
		destinationFile = filepath.Join(fileInformation.DestinationExtracted, masterFile)

		hash, err := SharedModels.CalculateMD5(destinationFile, 1025)
		if err != nil {
			return cstmerr.NewProcessError(cstmerr.PROCESS_HASH_ERROR, err)
		}
		localMovie.Link.FileHash = hex.EncodeToString(hash)
		localMovie.Link.PlayLink = filepath.Join(
			fileInformation.FileNameWithPrefix[0:len(fileInformation.FileNameWithPrefix)-4], masterFile)
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

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
		defer cancel()

		err = dbConnection.Save(ctx, &localMovie)
		if err != nil {
			return cstmerr.NewProcessError("failed to create slider", err)
		}

	} else {
		panic("not implemented")

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
		panic("not implemented")

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
		panic("not implemented")

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
		panic("not implemented")

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
		panic("not implemented")

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
		panic("not implemented")

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
		panic("not implemented")

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
		panic("not implemented")

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
