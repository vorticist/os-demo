package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/gorilla/mux"
	"github.com/vorticist/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type message struct {
	Message  string `json:"message"`
	LogLine  string `json:"log_line"`
	VideoURL string `json:"video_url"`
}

func main() {
	router := mux.NewRouter()

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./static/"))))
	router.HandleFunc("/detect", detectHandler).Methods("GET")
	router.HandleFunc("/", makeIndexHandler()).Methods("GET")

	fmt.Println("Server is running on :8080")
	http.ListenAndServe(":8080", router)
}

func makeIndexHandler() func(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("index...")
		tmpl.Execute(w, "index")
	}
}

// yolo predict model=yolov8n-seg.pt source='https://youtu.be/c8XQp5brszI' imgsz=320
func detectHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("failed to upgrade connection: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	for {
		msg := message{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("error %v", err)
			}
			break
		}

		url := msg.Message
		if len(url) == 0 {
			continue
		}

		logger.Infof("got url: %v", url)

		args := []string{
			"detect",
			"predict",
			"model='/server/best.pt'",
			fmt.Sprintf("source='%v'", url),
			"conf=0.70",
			"imgsz=640",
		}
		err = executeCommandWithOutputLogs(conn, "yolo", "/usr/src/ultralytics", args)
		if err != nil {
			logger.Error("error executing command: %v", err)
			return
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte("DONE."))
		if err != nil {
			logger.Error("error writing message: %v", err)
			return
		}
		mp4Path := "./static/output.mp4"
		aviPath, err := findVideoFile()
		if err != nil {
			logger.Error("error finding video file: %v", err)
			return
		}
		logger.Infof("found video file: %v", aviPath)

		args = []string{"-i", aviPath, "-vcodec", "libx264", "-vprofile", "high", "-crf", "28", mp4Path}
		err = executeCommand("ffmpeg", "/server", args)
		if err != nil {
			logger.Error("error executing command: %v", err)
			return
		}

		resultFileName := filepath.Base(mp4Path)
		msg = message{VideoURL: fmt.Sprintf("/static/%v", resultFileName)}
		err = conn.WriteMessage(websocket.TextMessage, getTemplate("templates/video.html", msg))
		if err != nil {
			logger.Error("error writing message: %v", err)
			return
		}
	}
}

func getTemplate(templatePath string, msg message) []byte {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		logger.Errorf("template failed to parse: %v", err)
		return nil
	}
	var renderedMessage bytes.Buffer
	err = tmpl.Execute(&renderedMessage, msg)
	if err != nil {
		logger.Errorf("template execution failed: %v", err)
		return nil
	}

	return renderedMessage.Bytes()
}

func executeCommand(command, cmdDir string, c []string) error {
	cmd := exec.Command(command, c...)
	cmd.Dir = cmdDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error reading stdout: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error reading stdout: %v", err)
	}
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %v", err)
	}
	multi := io.MultiReader(stdout, stderr)
	in := bufio.NewScanner(multi)
	for in.Scan() {
		line := in.Text()
		logger.Info(line)
	}

	if err := in.Err(); err != nil {
		return fmt.Errorf("error reading stdout: %v", err)
	}

	return nil
}

func executeCommandWithOutputLogs(conn *websocket.Conn, command, cmdDir string, c []string) error {
	cmd := exec.Command(command, c...)
	cmd.Dir = cmdDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error reading stdout: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error reading stdout: %v", err)
	}
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %v", err)
	}
	multi := io.MultiReader(stdout, stderr)
	in := bufio.NewScanner(multi)
	for in.Scan() {
		line := in.Text()
		logger.Info(line)
		if strings.Contains(line, "WARNING") {
			continue
		}
		msg := message{LogLine: line}
		err := conn.WriteMessage(websocket.TextMessage, getTemplate("templates/log.html", msg))
		if err != nil {
			break
		}
	}

	if err := in.Err(); err != nil {
		return fmt.Errorf("error reading stdout: %v", err)
	}

	return nil
}

func findVideoFile() (string, error) {
	folderPath := "/usr/src/ultralytics/runs/detect/predict"
	outputPath := ""
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Errorf("error walking: %v", err)
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".avi") {
			logger.Infof("found output file: %v", path)
			outputPath = path
			return filepath.SkipDir
		}

		return nil
	})
	if err != nil {
		logger.Errorf("error finding video file: %v", err)
		return "", err
	}
	return outputPath, nil
}

func moveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}
