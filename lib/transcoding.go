package hypnotic

import "mime/multipart"
import "os"
import "os/exec"
import "path"

const MAX_CONCURRENT_TRANSCODING_JOBS = 2
const TRANSCODING_SRC_DIRECTORY = "/home/regis/bazar/transcoding/src/"
const TRANSCODING_TMP_DIRECTORY = "/home/regis/bazar/transcoding/tmp/"
const TRANSCODING_DST_DIRECTORY = "/home/regis/bazar/transcoding/dst/"

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

func Transcode(file multipart.File, videoID string, filename string) {
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

	// Save to source directory
	srcFilePath := SaveToSrcDirectory(file, videoID, filename)
	defer os.Remove(srcFilePath)

	// Run transcoding and save to tmp directory
	convertedFilePath := RunTranscoding(video, job, srcFilePath)
	defer os.Remove(convertedFilePath)

	// Save to dst directory
	dstFilePath := TranscodedVideoPath(video.ID)
	os.Rename(convertedFilePath, dstFilePath)

	// Update job and video properties
	Db().Model(&job).Update("Status", "SUCCESS")
	Db().Model(&video).Update("Published", true)
}

func SaveToSrcDirectory(file multipart.File, videoID string, filename string) string {
	// Save file to temporary directory
	srcFilePath := path.Join(TRANSCODING_SRC_DIRECTORY, videoID+path.Ext(filename))
	tmpFile, err := os.Create(srcFilePath)
	check(err) // TODO mark job as failed
	defer tmpFile.Close()
	buffer := make([]byte, 1024)

	length := 0
	for err = nil; err == nil; {
		length, err = file.Read(buffer)
		tmpFile.Write(buffer[:length])
	}
	return srcFilePath
}

func RunTranscoding(video Video, job TranscodingJob, srcFilePath string) string {
	convertedFilePath := VideoPath(TRANSCODING_TMP_DIRECTORY, video.ID)
	stdout, stderr := exec.Command("avconv",
		"-i", srcFilePath,
		"-acodec", "mp3", "-vcodec", "libx264",
		"-y",
		convertedFilePath).Output()
	Db().Model(&job).Update("Stdout", stdout).Update("Stderr", stderr)
	return convertedFilePath
}
