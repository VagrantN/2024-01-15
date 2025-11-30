package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jung-kurt/gofpdf"
)

// Структура для входящего запроса
type CheckRequest struct {
	Links []string `json:"links"`
}

// Структура для хранения одной ссылки с статусом
type LinkStatus struct {
	URL    string `json:"url"`
	Status string `json:"status"`
}

// Структура для хранения одного запроса
type SavedRequest struct {
	ID    int          `json:"id"`
	Links []LinkStatus `json:"links"` // массив с ссылками и их статусами
}

// Главное хранилище
type Storage struct {
	Requests []SavedRequest `json:"requests"`
	NextID   int            `json:"next_id"`
}

type ReportRequest struct {
	LinksList []int `json:"links_list"`
}

var storage Storage

const dataFile = "F:\\Golang\\data.json"

// Загрузка данных из файла
func loadFromFile() Storage {
	file, err := os.Open(dataFile)
	if err != nil {
		fmt.Println("Файл не найден, начинаем с чистого листа")
		return Storage{
			Requests: []SavedRequest{},
			NextID:   1,
		}
	}
	defer file.Close()

	var loadedStorage Storage
	err = json.NewDecoder(file).Decode(&loadedStorage)
	if err != nil {
		log.Fatal("Ошибка чтения файла:", err)
	}

	fmt.Printf("Загружено: %d запросов\n", len(loadedStorage.Requests))
	return loadedStorage
}

// Сохранение данных в файл
func saveToFile() {
	file, err := os.Create(dataFile)
	if err != nil {
		log.Println("Ошибка создания файла:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(storage)
	if err != nil {
		log.Println("Ошибка записи:", err)
		return
	}

	fmt.Println("Данные сохранены")
}

func main() {
	fmt.Println("Запуск сервера...")

	// Загружаем данные при старте
	storage = loadFromFile()

	fmt.Printf("Следующий ID: %d\n", storage.NextID)

	// Регистрируем обработчики
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Server is running")
	})

	http.HandleFunc("/check-links", checkLinksHandler)
	http.HandleFunc("/generate-report", generateReportHandler)

	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Обработчик для проверки ссылок
func checkLinksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Только POST запросы", http.StatusMethodNotAllowed)
		return
	}

	// Парсим JSON из тела запроса
	var request CheckRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("Получены ссылки: %v\n", request.Links)

	// Создаем новый запрос
	newRequest := SavedRequest{
		ID: storage.NextID,
	}

	// Проверяем каждую ссылку и сохраняем статус для каждой
	for i := 0; i < len(request.Links); i++ {
		link := request.Links[i]
		status := checkLinkAvailability(link)

		// Сохраняем ссылку с ее статусом
		newRequest.Links = append(newRequest.Links, LinkStatus{
			URL:    link,
			Status: status,
		})

		fmt.Printf("Проверка %s: %s\n", link, status)
	}

	// Сохраняем в хранилище
	storage.Requests = append(storage.Requests, newRequest)
	storage.NextID++

	// Сохраняем в файл
	saveToFile()

	// Формируем ответ как в ТЗ - с мапой статусов для каждой ссылки
	linksMap := make(map[string]string)
	for _, linkStatus := range newRequest.Links {
		linksMap[linkStatus.URL] = linkStatus.Status
	}

	response := map[string]interface{}{
		"links":     linksMap,
		"links_num": newRequest.ID,
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	fmt.Printf("Обработан запрос #%d\n", newRequest.ID)
}

// Упрощенная проверка ссылки
func checkLinkAvailability(link string) string {
	// Просто пробуем сделать запрос
	resp, err := http.Get("http://" + link)
	if err != nil {
		return "not available"
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return "available"
	}
	return "not available"
}

func generateReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Только POST запросы", http.StatusMethodNotAllowed)
		return
	}

	// Парсим запрос
	var request ReportRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("Запрошен отчет для номеров: %v\n", request.LinksList)

	// Собираем все ссылки из запрошенных номеров
	var allLinks []LinkStatus
	for _, num := range request.LinksList {
		for _, savedReq := range storage.Requests {
			if savedReq.ID == num {
				// Добавляем все ссылки из этого запроса
				allLinks = append(allLinks, savedReq.Links...)
			}
		}
	}

	// Генерируем PDF
	pdf := generatePDF(allLinks)

	// Отправляем PDF как файл
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=report.pdf")
	w.Write(pdf)

	fmt.Println("PDF отчет отправлен")
}

func generatePDF(links []LinkStatus) []byte {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Links Report")
	pdf.Ln(20)

	pdf.SetFont("Arial", "", 12)
	for i := 0; i < len(links); i++ {
		link := links[i]
		pdf.Cell(0, 10, fmt.Sprintf("%s - %s", link.URL, link.Status))
		pdf.Ln(8)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		log.Println("Ошибка генерации PDF:", err)
		return []byte{}
	}
	return buf.Bytes()
}
