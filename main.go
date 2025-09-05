package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Note struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   []Content `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Content struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// 用于防止并发写入文件
var saveMutex sync.Mutex

func main() {
	// 检查并创建默认的data.txt和data_readable.txt文件（如果不存在）
	createDefaultFilesIfNotExists()
	
	http.HandleFunc("/save", saveHandler)
	http.HandleFunc("/load", loadHandler)
	http.HandleFunc("/ping", pingHandler) // 添加ping路由
	http.Handle("/", http.FileServer(http.Dir("./web/")))

	fmt.Println("Server started at http://localhost:1916")
	log.Fatal(http.ListenAndServe(":1916", nil))
}

// 检查并创建默认的data.txt和data_readable.txt文件（如果不存在）
func createDefaultFilesIfNotExists() {
	// 检查data.txt是否存在
	if _, err := os.Stat("data.txt"); os.IsNotExist(err) {
		// 创建空的笔记数组
		emptyNotes := []Note{}
		data, _ := json.Marshal(emptyNotes)
		
		// 写入data.txt
		err = ioutil.WriteFile("data.txt", data, 0644)
		if err != nil {
			log.Printf("Error creating default data.txt: %v", err)
		} else {
			log.Println("Created default data.txt")
		}
		
		// 写入格式化的data_readable.txt
		readableData, _ := json.MarshalIndent(emptyNotes, "", "  ")
		err = ioutil.WriteFile("data_readable.txt", readableData, 0644)
		if err != nil {
			log.Printf("Error creating default data_readable.txt: %v", err)
		} else {
			log.Println("Created default data_readable.txt")
		}
	} else {
		log.Println("data.txt already exists, skipping default file creation")
	}
}

// ping处理函数，用于检查连接状态
func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "ok", "message": "Server is running"}
	json.NewEncoder(w).Encode(response)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// 使用互斥锁防止并发写入
	saveMutex.Lock()
	defer saveMutex.Unlock()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// 保存到 data.txt
	err = ioutil.WriteFile("data.txt", body, 0644)
	if err != nil {
		http.Error(w, "Error saving to data.txt: " + err.Error(), http.StatusInternalServerError)
		return
	}

	// 同时保存为可读格式
	var notes []Note
	if err := json.Unmarshal(body, &notes); err == nil {
		readableData, _ := json.MarshalIndent(notes, "", "  ")
		err = ioutil.WriteFile("data_readable.txt", readableData, 0644)
		if err != nil {
			log.Printf("Error saving to data_readable.txt: %v", err)
		} else {
			log.Println("Data auto-saved to data.txt and data_readable.txt")
		}
	} else {
		log.Printf("Error parsing JSON data: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "success", "message": "Data saved successfully"}
	json.NewEncoder(w).Encode(response)
}

func loadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat("data.txt"); os.IsNotExist(err) {
		http.Error(w, "No saved data found", http.StatusNotFound)
		return
	}

	data, err := ioutil.ReadFile("data.txt")
	if err != nil {
		http.Error(w, "Error reading saved data: " + err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}