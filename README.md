# Веб-сервер для проверки доступности ссылок

## Описание
Простой веб-сервер на Go для проверки доступности интернет-ресурсов. Позволяет отправлять списки ссылок и получать отчеты в PDF формате.

## Функциональность
- **Проверка ссылок**: Принимает список URL и проверяет их доступность
- **Сохранение истории**: Каждому запросу присваивается уникальный номер
- **PDF отчеты**: Генерация отчетов по номерам ранее проверенных ссылок
- **Устойчивость к перезагрузке**: Данные сохраняются в файл и восстанавливаются при перезапуске

## Как использовать

### 1. Проверка ссылок
Отправьте POST запрос на `/check-links`:

**Запрос:**
```json
{
  "links": ["google.com", "yandex.ru", "invalid-site.xyz"]
}

{
  "links": {
    "google.com": "available",
    "yandex.ru": "available",
    "invalid-site.xyz": "not available"
  },
  "links_num": 1
}

{
  "links_list": [1, 2, 3]
}



go run start.go

http.HandleFunc("/test-pdf", func(w http.ResponseWriter, r *http.Request) {
    allLinks := []LinkStatus{
        {URL: "google.com", Status: "available"},
        {URL: "yandex.ru", Status: "available"},
        {URL: "invalid.gg", Status: "not available"},
    }
    pdf := generatePDF(allLinks)
    w.Header().Set("Content-Type", "application/pdf")
    w.Header().Set("Content-Disposition", "attachment; filename=report.pdf")
    w.Write(pdf)
})
