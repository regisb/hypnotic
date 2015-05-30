package hypnotic

import "fmt"
import "math/rand"
import "path"
import "time"

func TranscodedVideoPath(videoID string) string {
	return VideoPath(TRANSCODING_DST_DIRECTORY, videoID)
}

func VideoPath(directory string, videoID string) string {
	return path.Join(directory, videoID+".mp4")
}

func TranscodedVideoName(originalFileName string) string {
	originalExt := path.Ext(originalFileName)
	fileName := originalFileName[:len(originalFileName)-len(originalExt)]
	return fileName + ".mp4"
}

func RandomVideoID() string {
	rand.Seed(time.Now().UTC().UnixNano())
	videoIDRunes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var videoID = make([]uint8, VIDEO_ID_LENGTH)
	for index := 0; index < VIDEO_ID_LENGTH; index++ {
		videoID[index] = videoIDRunes[rand.Intn(len(videoIDRunes))]
	}
	return string(videoID)
}

func check(err error) {
	if err != nil {
		fmt.Println("panic ##########", err)
		panic(err)
	}
}
