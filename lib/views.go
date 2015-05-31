package hypnotic

import "encoding/json"
import "fmt"
import "html/template"
import "net/http"
import "os"
import "time"
import "github.com/gorilla/mux"

const (
	USER_DOES_NOT_EXIST = "USER_DOES_NOT_EXIST"
	EMPTY_USER_ID       = "EMPTY_USER_ID"
	UNPUBLISHED_VIDEO   = "UNPUBLISHED_VIDEO"
)

func HandleRoutes() {
	router := mux.NewRouter()
	router.HandleFunc("/users", UsersHandler).Methods("GET")
	router.HandleFunc("/user/{id:[a-zA-Z0-9]+}", GetUserHandler).Methods("GET")
	router.HandleFunc("/user", PostUserHandler).Methods("POST")
	router.HandleFunc("/jobs", JobsHandler).Methods("GET")
	router.HandleFunc("/videos", VideosHandler).Methods("GET")
	router.HandleFunc("/video", PostVideoHandler).Methods("POST")
	router.HandleFunc("/video/{id:[a-zA-Z0-9]+}", DeleteVideoHandler).Methods("DELETE")
	router.HandleFunc("/{id:[a-zA-Z0-9]+}.mp4", VideoHandler).Methods("GET")
	router.HandleFunc("/{id:[a-zA-Z0-9]+}", HtmlVideoHandler).Methods("GET")
	http.Handle("/", router)
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	var users []User
	Db().Find(&users)

	var userIDs = make([]string, len(users))
	for index, user := range users {
		userIDs[index] = user.ID
	}
	var response = map[string]interface{}{
		"count": len(users),
		"users": userIDs,
	}
	JsonMarshal(w, response)
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	var users []User
	Db().Where("ID = ?", mux.Vars(r)["id"]).Find(&users)

	if len(users) == 0 {
		JsonError(w, USER_DOES_NOT_EXIST)
	} else {
		var response = map[string]string{
			"id": users[0].ID,
		}
		JsonMarshal(w, response)
	}
}

func PostUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.FormValue("id")
	if len(userID) == 0 {
		JsonError(w, EMPTY_USER_ID)
		return
	}
	user := User{
		ID: userID,
	}
	var response = map[string]interface{}{}
	err := Db().Create(&user).Error
	if err != nil {
		JsonMarshalError(w, response, err.Error())
		return
	} else {
		response["id"] = userID
		JsonMarshal(w, response)
	}
}

func JobsHandler(w http.ResponseWriter, r *http.Request) {
	var jobs []TranscodingJob
	Db().Find(&jobs)

	var jobArr = make([]map[string]interface{}, len(jobs))
	for index, job := range jobs {
		jobArr[index] = map[string]interface{}{
			"id":         job.ID,
			"video_id":   job.VideoID,
			"status":     job.Status,
			"created_at": job.CreatedAt.String(),
			"updated_at": job.UpdatedAt.String(),
		}
	}
	var response = map[string]interface{}{
		"count": len(jobs),
		"jobs":  jobArr,
	}
	JsonMarshal(w, response)
}

func PostVideoHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	userID := r.FormValue("user_id")
	if err != nil {
		JsonError(w, err.Error())
		return
	}

	var users []User
	Db().Where("ID = ?", userID).Find(&users)
	if len(users) == 0 {
		JsonError(w, USER_DOES_NOT_EXIST)
		return
	}

	videoID := RandomVideoID()
	var video = Video{
		OriginalFilename: header.Filename,
		ID:               videoID,
		UserID:           userID,
	}
	err = Db().Create(&video).Error
	if err != nil {
		JsonError(w, err.Error())
		return
	}

	var response = map[string]interface{}{
		"id":                videoID,
		"original_filename": header.Filename,
		"user_id":           userID,
	}
	go Transcode(file, videoID, header.Filename)
	JsonMarshal(w, response)
}

func DeleteVideoHandler(w http.ResponseWriter, r *http.Request) {
	videoID := mux.Vars(r)["id"]

	// TODO can't delete a currently-running job

	// Delete file
	os.Remove(TranscodedVideoPath(videoID))

	// Delete db entries
	Db().Where("ID = ?", videoID).Delete(Video{})
	Db().Where("video_id = ?", videoID).Delete(TranscodingJob{})
}

func VideosHandler(w http.ResponseWriter, r *http.Request) {
	var videos []Video
	Db().Find(&videos)

	responseVideos := make([]map[string]string, len(videos))
	for index, video := range videos {
		responseVideos[index] = map[string]string{
			"id":                video.ID,
			"original_filename": video.OriginalFilename,
			"user_id":           video.UserID,
		}
	}
	response := map[string]interface{}{
		"videos": responseVideos,
	}
	JsonMarshal(w, response)
}

func VideoHandler(w http.ResponseWriter, r *http.Request) {
	video := GetPublishedVideoOr404(w, r)
	if video == nil {
		return
	}

	videoFile, err := os.Open(TranscodedVideoPath(video.ID))
	check(err)
	defer videoFile.Close()
	// TODO pass correct time value
	http.ServeContent(w, r, TranscodedVideoName(video.OriginalFilename), time.Time{}, videoFile)
}

func HtmlVideoHandler(w http.ResponseWriter, r *http.Request) {
	video := GetVideoOr404(w, r)
	if video == nil {
		return
	}

	t, err := template.ParseFiles("templates/video.html")
	check(err)
	t.Execute(w, video)
}

func GetPublishedVideoOr404(w http.ResponseWriter, r *http.Request) *Video {
	video := GetVideoOr404(w, r)
	if video != nil && !video.Published {
		JsonError(w, UNPUBLISHED_VIDEO)
		return nil
	}
	return video
}

func GetVideoOr404(w http.ResponseWriter, r *http.Request) *Video {
	var video Video
	videoId := mux.Vars(r)["id"]
	if Db().Where("ID = ?", videoId).First(&video).RecordNotFound() {
		http.NotFound(w, r)
		return nil
	}
	return &video
}

func JsonError(w http.ResponseWriter, err string) error {
	response := map[string]interface{}{}
	return JsonMarshalError(w, response, err)
}

func JsonMarshalError(w http.ResponseWriter, data map[string]interface{}, err string) error {
	data["error"] = err
	return JsonMarshal(w, data)
}

func JsonMarshal(w http.ResponseWriter, data interface{}) error {
	dump, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println("Error while opening parsing JSON:", err)
	}
	w.Write(dump)
	return err
}
