package main

import (
  "fmt"
  "html/template"
  "log"
  "net/http"
  "os"
  "github.com/skip2/go-qrcode"
)

func main() {
  http.HandleFunc("/", formHandler)
  http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

  fmt.Println("Starting server on :8080...")
  if err := http.ListenAndServe(":8080", nil); err != nil {
    log.Fatal("Server failed to start:", err)
  }
}

type FormData struct {
  FirstName string
  SurName   string
  Company   string
  Title     string
  Email     string
  Phone     string
  QRCode    string 
}

func formHandler(w http.ResponseWriter, r *http.Request) {
  tmpl, err := template.ParseFiles("templates/form.html")
  if err != nil {
    log.Println("Error loading form:", err)
    http.Error(w, "Could not load form: "+err.Error(), http.StatusInternalServerError)
    return
  }

  formData := FormData{} 

  if r.Method == http.MethodPost {
    formData = FormData{
      FirstName: r.FormValue("firstName"),
      SurName:   r.FormValue("surName"),
      Company:   r.FormValue("company"),
      Title:     r.FormValue("title"),
      Email:     r.FormValue("email"),
      Phone:     r.FormValue("phone"),
    }

    vCard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nN:%s;%s;;;\nFN:%s %s\nORG:%s\nTITLE:%s\nEMAIL:%s\nTEL:%s\nEND:VCARD",
    formData.SurName, formData.FirstName, formData.FirstName, formData.SurName, formData.Company, formData.Title, formData.Email, formData.Phone)


    log.Println("Generated vCard content:", vCard)


    if _, err := os.Stat("static"); os.IsNotExist(err) {
      err := os.Mkdir("static", 0755)
      if err != nil {
        log.Println("Error creating static directory:", err)
        http.Error(w, "Could not create static directory", http.StatusInternalServerError)
        return
      }
    }

    filePath := fmt.Sprintf("static/%s_%s_qrcode.png", formData.FirstName, formData.SurName)
    err = qrcode.WriteFile(vCard, qrcode.Medium, 256, filePath)
    if err != nil {
      log.Println("Error generating QR code:", err)
      http.Error(w, "Could not generate QR code", http.StatusInternalServerError)
      return
    }

    formData.QRCode = "/" + filePath 
    log.Println("QR code saved as:", formData.QRCode)
  }

  if err := tmpl.Execute(w, formData); err != nil {
    log.Println("Error rendering template:", err)
    http.Error(w, "Could not render template", http.StatusInternalServerError)
  }
}

