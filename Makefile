all: build
migrate: build
	./hypnotic -migrate
build:
	go build
requirements: 
	go get github.com/gorilla/mux
	go get github.com/jinzhu/gorm
	go get github.com/mattn/go-sqlite3
