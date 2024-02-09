package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vorticist/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	router := mux.NewRouter()

	router.PathPrefix("/client/").Handler(http.StripPrefix("/client", http.FileServer(http.Dir("./client/"))))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./static/"))))
	router.HandleFunc("/predict", PredictHandler).Methods("GET")

	fmt.Println("Server is running on :8080")
	http.ListenAndServe(":8080", router)
}

// yolo predict model=yolov8n-seg.pt source='https://youtu.be/J-PuuKF3xyk' imgsz=320
func PredictHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	url := params.Get("url")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("failed to upgrade connection: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	logger.Infof("got url: %v", url)

	args := []string{
		"detect",
		"predict",
		"model=yolov8n.pt",
		fmt.Sprintf("source='%v'", url),
		"imgsz=320",
	}
	err = ExecuteCommandWithOutputLogs(conn, "yolo", "/usr/src/ultralytics", args)
	if err != nil {
		logger.Error("error executing command: %v", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, []byte("DONE."))
	if err != nil {
		logger.Error("error writing message: %v", err)
		return
	}
	m3u8Path := "./static/output.m3u8"
	aviPath, err := findVideoFile()
	if err != nil {
		logger.Error("error finding video file: %v", err)
		return
	}
	logger.Infof("found video file: %v", aviPath)

	args = []string{
		"-i",
		aviPath,
		"-c:a",
		"copy",
		"-f",
		"hls",
		"-hls_playlist_type",
		"vod",
		m3u8Path,
	}
	err = ExecuteCommand("ffmpeg", "/server", args)
	if err != nil {
		logger.Error("error executing command: %v", err)
		return
	}

	resultFileName := filepath.Base(m3u8Path)
	err = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("/static/%v", resultFileName)))
	if err != nil {
		logger.Error("error writing message: %v", err)
		return
	}
	//file, err := os.Open(aviPath)
	//if err != nil {
	//	logger.Errorf("error opening file: %s", err)
	//	return
	//}
	//defer file.Close()
	//
	//buffer := make([]byte, 1024)
	//for {
	//	_, err := file.Read(buffer)
	//	if err != nil {
	//		logger.Error("error reading: %v", err)
	//		break
	//	}
	//	err = conn.WriteMessage(websocket.BinaryMessage, buffer)
	//	if err != nil {
	//		logger.Errorf("error writing: %v", err)
	//		break
	//	}
	//}
	//
	//logger.Printf("finished streaming video: %s", aviPath)
}
func ExecuteCommand(command, cmdDir string, c []string) error {
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

func ExecuteCommandWithOutputLogs(conn *websocket.Conn, command, cmdDir string, c []string) error {
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
		err := conn.WriteMessage(websocket.TextMessage, []byte(line))
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
