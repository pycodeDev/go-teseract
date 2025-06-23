package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/otiai10/gosseract/v2"
)

type OCRSuccessResponse struct {
	Status  bool   `json:"status"`
	Msg     string `json:"msg"`
	OCRData string `json:"ocr_data"`
}

type OCRErrorResponse struct {
	Status bool   `json:"status"`
	ErrMsg string `json:"err_msg"`
}

func main() {
	http.HandleFunc("/ocr", handleOCR)
	fmt.Println("🚀 OCR server running on :8003")
	http.ListenAndServe(":8003", nil)
}

func handleOCR(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Only POST allowed")
		return
	}

	var imageBytes []byte
	var url string

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		var req struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
			return
		}
		url = req.URL
	} else {
		url = r.FormValue("url")
	}

	if url != "" {
		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			writeError(w, http.StatusBadRequest, "Failed to download image from URL")
			return
		}
		defer resp.Body.Close()
		imageBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to read image from URL")
			return
		}
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		file, _, err := r.FormFile("image")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Failed to get image file: "+err.Error())
			return
		}
		defer file.Close()
		imageBytes, err = io.ReadAll(file)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to read image file: "+err.Error())
			return
		}
	} else {
		writeError(w, http.StatusBadRequest, "No image source provided")
		return
	}

	client := gosseract.NewClient()
	defer client.Close()

	if err := client.SetImageFromBytes(imageBytes); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to set image: "+err.Error())
		return
	}

	text, err := client.Text()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "OCR error: "+err.Error())
		return
	}

	writeSuccess(w, text)
}

func writeSuccess(w http.ResponseWriter, ocrText string) {
	w.WriteHeader(http.StatusOK)
	resp := OCRSuccessResponse{
		Status:  true,
		Msg:     "OCR berhasil",
		OCRData: ocrText,
	}
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, statusCode int, errMsg string) {
	w.WriteHeader(statusCode)
	resp := OCRErrorResponse{
		Status: false,
		ErrMsg: errMsg,
	}
	json.NewEncoder(w).Encode(resp)
}
