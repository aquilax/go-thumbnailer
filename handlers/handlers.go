package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/pydima/go-thumbnailer/image"
	"github.com/pydima/go-thumbnailer/models"
	"github.com/pydima/go-thumbnailer/utils"
)

type Task struct {
	Path       string
	Identifier string
	Delay      bool
	TaskID     string
}

func checkParams(t *Task) (err error) {
	if t.TaskID == "" {
		t.TaskID = utils.UUID()
	}
	return
}

func CreateThumbnail(w http.ResponseWriter, r *http.Request) {
	var t Task
	d := json.NewDecoder(r.Body)

	if err := d.Decode(&t); err != nil {
		os.Exit(1)
	}

	if !t.Delay {
		if err := checkParams(&t); err == nil {
			fmt.Fprintf(w, "OK")
		} else {
			log.Fatal("Good buy.")
		}
	}

	db_i := models.Image{
		OriginalPath: t.Path,
		Identifier:   t.Identifier,
	}

	if db_i.Exist() {
		fmt.Println("This image is already exist.")
		return
	}

	var i io.ReadCloser
	if t.Path[:4] == "http" {
		i, _ = utils.DownloadImage(t.Path)
	} else {
		i, _ = utils.ReadImage(t.Path)
	}

	path, err := image.ProcessImage(i)
	if err != nil {
		log.Fatal("Sorry.")
	}

	db_i.Path = path

	models.Db.Create(&db_i)

}
