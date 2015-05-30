all: build

build:
	go build behmo.com/hypnotic
requirements: 
	go get github.com/gorilla/mux
	go get github.com/jinzhu/gorm
	go get github.com/mattn/go-sqlite3
