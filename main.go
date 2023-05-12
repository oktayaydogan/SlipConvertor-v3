package main

import (
	"context"
	"errors"
	"log"
	"regexp"
	"strings"
	"os"
	"fmt"
	"io/ioutil"

	vision "cloud.google.com/go/vision/apiv1"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"google.golang.org/api/option"
)

var convertButton *widget.Button
var openTxtButton *widget.Button

func main() {
	myApp := app.New()
	//new window with title "Slip Converter"
	myWindow := myApp.NewWindow("Slip Converter")

	//windows size
	myWindow.Resize(fyne.NewSize(600, 400))

	selectedRegex := ""
	selectedFolder := ""

	//select regex label
	typeLabel := widget.NewLabel("Slip Türü Seçiniz:")

	//select regex button
	options := []string{"Steam", "Netflix"}
	typeSelect := widget.NewSelect(options, func(selected string) {
		selectedRegex = selected
	})

	// Seçilen klasörün yolunu göstermek için bir etiket oluşturun
	folderSelectedLabel := widget.NewLabel("Seçilen Klasör:")
	folderSelectedBindLabel := widget.NewLabel("")

	folderSelectedLabel.Hide()
	folderSelectedBindLabel.Hide()

	folderLabel := widget.NewLabel("Slip Klasörü Seçiniz:")
	// Klasör seçme işlemi başarılı olduğunda çağrılacak fonksiyon
	onFolderSelected := func(folder fyne.ListableURI, err error) {
		if err == nil {
			// global bir değişkene seçilen klasörün yolunu atayın
			selectedFolder = folder.Path()

			folderSelectedBindLabel.SetText(selectedFolder)
			folderSelectedLabel.Show()
			folderSelectedBindLabel.Show()
		}
	}

	// "SelectFileDialog" widget'ını oluşturun
	folderSelect := dialog.NewFolderOpen(onFolderSelected, myWindow)

	// "SelectFileDialog" widget'ını çalıştırmak için bir buton oluşturun
	folderSelectButton := widget.NewButton("Klasör Seç", func() {
		folderSelect.Show()
	})


	// LOG Textarea
	logLabel := widget.NewLabel("Log:")
	logText := widget.NewMultiLineEntry()
	logText.Disable()
	logLabel.Hide()
	logText.Hide()

	openTxtLabel := widget.NewLabel("")

	// txt dosyasını aç butonu
	openTxtButton := widget.NewButton("txt dosyasını aç", func() {
		os.StartProcess("notepad.exe", []string{"codes.txt"}, nil)
	})

	openTxtLabel.Hide()
	openTxtButton.Hide()

	convertLabel := widget.NewLabel("")
	// çevir butonu
	convertButton := widget.NewButton("Çevir", func() {
		// convert fonksiyonunu çağırın ve seçilen slip türünü ve klasörün yolunu, pencereyi ve çevir butonunu ve çevir etiketini parametre olarak gönderin
		convert(selectedRegex, selectedFolder, myWindow, convertLabel, convertButton, logText, logLabel, openTxtLabel, openTxtButton)
	})

	// Grid layout

	grid := container.New(layout.NewFormLayout(), typeLabel, typeSelect, folderLabel, folderSelectButton, folderSelectedLabel, folderSelectedBindLabel, convertLabel, convertButton, logLabel, logText, openTxtLabel, openTxtButton)

	myWindow.SetContent(grid)
	myWindow.ShowAndRun()
}

func convert(slipType string, folder string, myWindow fyne.Window, convertLabel *widget.Label, convertButton *widget.Button, logText *widget.Entry, logLabel *widget.Label, openTxtLabel *widget.Label, openTxtButton *widget.Button){
	// slipType: Steam, Netflix 
	// folder: seçilen klasörün yolunu tutan değişken

	codesArray := []string{}

	// Seçilen slip türü veya klasör yoksa hata mesajı gösterin
	if slipType == "" {
		dialog.ShowError(errors.New("slip türü seçin"), myWindow)
		return
	}

	// Seçilen klasör yoksa hata mesajı gösterin
	if folder == "" {
		dialog.ShowError(errors.New("klasör seçin"), myWindow)
		return
	}

	// slip türüne göre regex'i tutan bir map oluşturun
	regexList := map[string]string{
		"Steam":   "[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}",
		"Netflix": "NA[A-Z0-9]{14}",
	}

	convertLabel.SetText("Çeviriliyor...")
	//convertButton.Disable()

	// regexList'ten seçilen slip türüne göre regex'i alın
	regex, err := regexp.Compile(regexList[slipType])
	if err != nil {
		log.Fatalf("regexp.Compile: %v", err)
	}

	// Google Vision API için bir istemci oluşturun
	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx, option.WithCredentialsFile("credentials.json"))

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Resimleri oku
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		log.Fatal(err)
	}

	// codes.txt dosyasını sil
	os.Remove("codes.txt")

	// codes.txt dosyası yoksa oluştur
	file, err := os.Create("codes.txt")

	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()
	
	logLabel.Show()
	logText.Show()
	// logText widget'ına başlangıç mesajı yazdırın
	logText.SetText("Çevirme işlemi başladı...\n")

	// logText widget'ının yüksekliğini ayarlayın
	logText.Resize(fyne.NewSize(600, 300))

	// Klasördeki tüm resimleri döngüye alın
	for _, f := range files {
		// Dosya uzantısı .jpg, .jpeg veya .png değilse döngüyü atlayın
		if !strings.HasSuffix(f.Name(), ".jpg") && !strings.HasSuffix(f.Name(), ".jpeg") && !strings.HasSuffix(f.Name(), ".png") {
			continue
		}

		// Dosya adını alın
		image_name := f.Name()

		// logText widget'ına dosya adını yeni bir satıra yazdırın
		logText.SetText("Dosya adı: "+image_name+"\n" + logText.Text)

		// Dosya yolunu oluşturun
		filePath := folder + "/" + image_name

		// Dosyayı oku
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		defer file.Close()

		// Dosyayı Google Vision API'ye gönderin
		image, err := vision.NewImageFromReader(file)
		if err != nil {
			log.Fatalf("Failed to create image: %v", err)
		}

		// OCR işlemi
		texts, err := client.DetectTexts(ctx, image, nil, 10)
		if err != nil {
			log.Fatalf("Failed to detect text: %v", err)
		}

		textsTrimmed := strings.ReplaceAll(texts[0].Description, " ", "")

		codes := regex.FindAllString(textsTrimmed, -1)
			fmt.Println(codes)
		for _, code := range codes {
			codesArray = append(codesArray, code + " - " + image_name)
		}
	}

	for _, code := range codesArray {
		// codes.txt dosyasına kodları yazdırın
		file.WriteString(code + "\n")
	}
	
	// Çevirme işlemi bittiğinde başarılı mesajı gösterin
	dialog.ShowInformation("Bitti", "Çevirme işlemi bitti", myWindow)

	logLabel.Hide()
	logText.Hide()

	openTxtLabel.Show()
	openTxtButton.Show()
}

