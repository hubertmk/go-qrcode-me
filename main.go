package main

import (
    "fmt"
    "github.com/skip2/go-qrcode"
    "html/template"
    "image"
    "image/color"
    "image/draw"
    "image/png"
    "log"
    "net/http"
    "os"
    "path/filepath"
)

const (
    wallpaperWidth  = 1170
    wallpaperHeight = 2532
    qrCodeSize      = 800
)

func main() {
    http.HandleFunc("/", formHandler)
    http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
        file := r.URL.Path[len("/static/"):]
        if file == "" || file == "/" {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        http.ServeFile(w, r, "static/"+file)
    })

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

        fileName := fmt.Sprintf("%s_%s_qrcode.png", formData.FirstName, formData.SurName)
        filePath := filepath.Join("static", fileName)

        qr, err := qrcode.New(vCard, qrcode.Medium)
        if err != nil {
            log.Println("Error generating QR code:", err)
            http.Error(w, "Could not generate QR code", http.StatusInternalServerError)
            return
        }
        qr.BackgroundColor = color.White
        qr.ForegroundColor = color.Black

        wallpaper := image.NewRGBA(image.Rect(0, 0, wallpaperWidth, wallpaperHeight))
        draw.Draw(wallpaper, wallpaper.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

        qrImage := qr.Image(qrCodeSize)
        offset := image.Pt((wallpaperWidth-qrCodeSize)/2, (wallpaperHeight-qrCodeSize)/2)
        draw.Draw(wallpaper, qrImage.Bounds().Add(offset), qrImage, image.Point{}, draw.Over)

        wallpaperFile, err := os.Create(filePath)
        if err != nil {
            log.Println("Error saving wallpaper:", err)
            http.Error(w, "Could not save wallpaper", http.StatusInternalServerError)
            return
        }
        defer wallpaperFile.Close()
        if err := png.Encode(wallpaperFile, wallpaper); err != nil {
            log.Println("Error encoding wallpaper to PNG:", err)
            http.Error(w, "Could not encode wallpaper as PNG", http.StatusInternalServerError)
            return
        }

        formData.QRCode = "/" + filePath 
        log.Println("Wallpaper saved as:", formData.QRCode)
    }

    if err := tmpl.Execute(w, formData); err != nil {
        log.Println("Error rendering template:", err)
        http.Error(w, "Could not render template", http.StatusInternalServerError)
    }
}

