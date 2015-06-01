package main

import "flag"
import "fmt"
import "net/http"

import "github.com/regisb/hypnotic/lib"

func main() {
	var serve = flag.Bool("serve", false, "Run the transcription server")
	var migrate = flag.Bool("migrate", false, "Migrate the database")
	var host = flag.String("host", "0.0.0.0:8079", "Host on which to run the server")
	flag.Parse()

	hypnotic.HandleRoutes()
	hypnotic.InitialiseTranscodingJobMessages()

	if *migrate {
		hypnotic.MigrateDb()
	}
	if *serve {
		fmt.Println("Serving on ", *host)
		http.ListenAndServe(*host, nil)
	}
}
