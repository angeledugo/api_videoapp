package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/angeledugo/video-app-api/database"
	"github.com/angeledugo/video-app-api/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	ung "github.com/dillonstreator/go-unique-name-generator"
)

func GetVideos(w http.ResponseWriter, r *http.Request) {

	var videos []models.Video
	database.DB.Find(&videos)
	json.NewEncoder(w).Encode(videos)
}

func UploadVideo(w http.ResponseWriter, r *http.Request) {

	fmt.Println("subiendo video")

	// Recibir el video desde la solicitud
	file, handler, err := r.FormFile("video")
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Generar un nombre único para el archivo
	defaultGenerator := ung.NewUniqueNameGenerator()
	filename := defaultGenerator.Generate()

	extension := filepath.Ext(handler.Filename)
	newFilename := filename + extension

	// Crear credenciales estáticas
	/*staticCreds := credentials.NewStaticCredentialsProvider(
		accessKeyID,
		secretAccessKey,
		"",
	)*/

	// Crear una configuración de AWS con credenciales estáticas
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(getRegionFromEnv()),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			""),
		),
	)

	//region := getRegionFromConfig(cfg)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	// Usar la configuración de AWS para interactuar con S3
	// ... (código para interactuar con S3 usando la configuración `cfg`)
	// Crear un cliente S3
	client := s3.NewFromConfig(cfg)
	// Subir el video al bucket de S3
	bucket := os.Getenv("S3_BUCKET_NAME")
	folder := os.Getenv("S3_BUCKET_FOLDER")
	if folder != "" {
		newFilename = folder + "/" + newFilename
	}

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(newFilename),
		Body:   file,
		ACL:    "public-read",
	})

	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	region := getRegionFromEnv()

	// Generar la URL del video
	videoURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, newFilename)

	video := models.Video{
		Title: newFilename,
		URL:   videoURL,
	}
	result := database.DB.Create(&video)

	if result.Error != nil {
		http.Error(w, "Error saving video to database: "+result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Enviar la respuesta con la información del video, incluyendo la URL
	json.NewEncoder(w).Encode(map[string]string{
		"url":  videoURL,
		"name": newFilename,
	})

}

func getRegionFromEnv() string {
	return os.Getenv("AWS_REGION")
}
