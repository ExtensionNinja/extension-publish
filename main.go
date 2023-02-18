package main

import (
	"context"
	"fmt"
	"log"
	"os"
)

func main() {
	action := os.Getenv("INPUT_ACTION")
	if action == "test" {
		writeOutput("result", "testOK")
		return
	}
	if action != "upload" && action != "uploadPublish" {
		log.Fatal("action not supported")
		return
	}

	clientID := os.Getenv("INPUT_CLIENTID")
	clientSecret := os.Getenv("INPUT_CLIENTSECRET")
	refreshToken := os.Getenv("INPUT_CLIENTREFRESHTOKEN")
	extensionID := os.Getenv("INPUT_EXTENSIONID")
	extensionFilePath := os.Getenv("INPUT_EXTENSIONFILE")

	file, err := os.ReadFile(extensionFilePath)
	if err != nil {
		fmt.Printf("Failed to read extension file. Err: %v", err)
		return
	}

	ctx := context.Background()

	token, err := refreshAccessToken(ctx, clientID, clientSecret, refreshToken)
	if err != nil {
		log.Fatalf("Failed to acquire access token. Err: %v", err)
		return
	}
	fmt.Println("Token refresh succeeded!")

	if action == "upload" || action == "uploadPublish" {
		uploadResult, err := uploadExtension(ctx, token, extensionID, file)
		if err != nil {
			log.Fatalf("Failed to upload: Err: %v", err)
			return
		}
		fmt.Printf("Upload result: %+v \n", uploadResult)
		if action == "upload" {
			writeOutput("result", "uploadOK")
		}
	}
	if action == "uploadPublish" {
		publishResult, err := publishExtension(ctx, token, extensionID)
		if err != nil {
			log.Fatalf("Failed to publish: Err: %v", err)
			return
		}
		fmt.Printf("Publish result: %+v \n", publishResult)
		writeOutput("result", "publishOK")
	}
}

func writeOutput(name string, value string) {
	file := os.Getenv("GITHUB_OUTPUT")
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.Write([]byte(name + "=" + value + "\n"))
	if err != nil {
		log.Fatal(err)
	}
}

