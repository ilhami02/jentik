package utils

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func AnalyzeImageWithGemini(file multipart.File, fileSize int64, mimeType string) (string, error) {
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash")

	imgData, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	prompt := genai.Text("Kamu adalah asisten AI untuk mendeteksi area rawan jentik nyamuk DBD. Perhatikan gambar ini baik-baik. Apakah terdapat genangan air, barang bekas penampung air, atau lingkungan yang berpotensi menjadi tempat berkembang biak nyamuk? Balas dengan format JSON yang rapi: {\"is_rawan\": true/false, \"alasan\": \"penjelasan singkat maksimal 2 kalimat\", \"saran\": \"tindakan yang harus dilakukan\"}")

	formatGambar := strings.TrimPrefix(mimeType, "image/")
	imgPart := genai.ImageData(formatGambar, imgData)

	resp, err := model.GenerateContent(ctx, prompt, imgPart)
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
	}

	return "", fmt.Errorf("tidak ada respon dari AI")
}