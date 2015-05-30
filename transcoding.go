package hypnotic

import "mime/multipart"
import "os"
import "os/exec"
import "path"

var TRANSCODING_JOB_MESSAGES = make(chan int, MAX_CONCURRENT_TRANSCODING_JOBS)

func InitialiseTranscodingJobMessages() {
	for index := 0; index < MAX_CONCURRENT_TRANSCODING_JOBS; index++ {
		TRANSCODING_JOB_MESSAGES <- index
	}
}

func GetTranscodingToken() int {
	index := <-TRANSCODING_JOB_MESSAGES
	return index
}

func FreeTranscodingToken(index int) {
	TRANSCODING_JOB_MESSAGES <- index
}

func Transcode(file multipart.File, videoID string, fileName string) {
	var job TranscodingJob
	var video Video

	Db().Where("ID = ?", videoID).First(&video)
	Db().Where(TranscodingJob{VideoID: videoID}).FirstOrCreate(&job)
	Db().Model(&video).Update("Published", false)

	// Get transcoding token
	Db().Model(&job).Update("Status", "WAITING")
	token := GetTranscodingToken()
	defer FreeTranscodingToken(token)
	Db().Model(&job).Update("Status", "TRANSCODING")

	// Save file to /tmp
	srcFilePath := path.Join(TRANSCODING_SRC_DIRECTORY, videoID+path.Ext(fileName))
	tmpFile, err := os.Create(srcFilePath)
	check(err) // TODO mark job as failed
	defer tmpFile.Close()
	buffer := make([]byte, 1024)

	length := 0
	for err = nil; err == nil; {
		length, err = file.Read(buffer)
		tmpFile.Write(buffer[:length])
	}

	// Remove original file
	defer os.Remove(srcFilePath)

	// Convert file
	convertedFilePath := VideoPath(TRANSCODING_TMP_DIRECTORY, video.ID)
	defer os.Remove(convertedFilePath)
	_, avconvErr := exec.Command("avconv",
		"-i", srcFilePath,
		"-acodec", "mp3", "-vcodec", "libx264",
		"-y",
		convertedFilePath).Output()
	if avconvErr != nil {

	}

	// Move mp4 file to transcoding destination
	dstFilePath := TranscodedVideoPath(video.ID)
	os.Rename(convertedFilePath, dstFilePath)

	// TODO add more steps to the transcoding job
	Db().Model(&job).Update("Status", "SUCCESS")
	Db().Model(&video).Update("Published", true)
}
