package hypnotic

import "flag"
import "fmt"
import "net/http"

const VIDEO_ID_LENGTH = 15
const MAX_CONCURRENT_TRANSCODING_JOBS = 2
const TRANSCODING_SRC_DIRECTORY = "/home/regis/bazar/transcoding/src/"
const TRANSCODING_TMP_DIRECTORY = "/home/regis/bazar/transcoding/tmp/"
const TRANSCODING_DST_DIRECTORY = "/home/regis/bazar/transcoding/dst/"
const (
	USER_DOES_NOT_EXIST = "USER_DOES_NOT_EXIST"
	EMPTY_USER_ID       = "EMPTY_USER_ID"
	UNPUBLISHED_VIDEO   = "UNPUBLISHED_VIDEO"
)

func main() {
	var serve = flag.Bool("serve", false, "Run the transcription server")
	var migrate = flag.Bool("migrate", false, "Migrate the database")
	var host = flag.String("host", "0.0.0.0:8079", "Host on which to run the server")
	flag.Parse()

	HandleRoutes()
	InitialiseTranscodingJobMessages()

	if *migrate {
		MigrateDb()
	}
	if *serve {
		fmt.Println("Serving on ", *host)
		http.ListenAndServe(*host, nil)
	}
}
