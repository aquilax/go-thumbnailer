package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pydima/go-thumbnailer/handlers"
)

func main() {
	var (
		image_source []handlers.ImageSource
		task         handlers.Task
	)

	image_source = append(image_source, handlers.ImageSource{"http://ecx.images-amazon.com/images/I/51eDwv7tCtL._SX442_BO1,204,203,200_.jpg", ""})
	task.Images = image_source
	task.Delay = false

	data, err := json.Marshal(task)
	if err != nil {
		os.Exit(1)
	}

	resp, err := http.Post("http://localhost:8080/thumbnail", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Get error ", err)
	} else {
		fmt.Println(resp.Status)
	}
}
