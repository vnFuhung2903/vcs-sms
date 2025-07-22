package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

const (
	baseURL = "http://localhost:8080"
	timeout = 30 * time.Second
)

var (
	httpClient  = &http.Client{Timeout: timeout}
	authToken   string
	containerID string
)

func getAuthTokenFromRedis() (string, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	key, err := rdb.Get("token").Result()
	if err != nil {
		return "", fmt.Errorf("failed to get Redis keys: %v", err)
	}

	return key, nil
}

func makeRequest(method, endpoint string, payload interface{}, useAuth bool, contentType string) (*http.Response, error) {
	var body io.Reader
	switch v := payload.(type) {
	case nil:
		body = nil
	case *bytes.Buffer:
		body = v
	default:
		jsonData, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+endpoint, body)
	if err != nil {
		return nil, err
	}

	if contentType == "" {
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Set("Content-Type", contentType)
	}

	if useAuth && authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	return httpClient.Do(req)
}

func printResponse(testName string, resp *http.Response, err error) {
	fmt.Printf("\n=== %s ===\n", testName)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("✅ Success!\n")
	} else {
		fmt.Printf("⚠️  Status: %d\n", resp.StatusCode)
	}
}

func TestSetupAuthToken(t *testing.T) {
	token, err := getAuthTokenFromRedis()
	if err != nil {
		authToken = ""
	} else {
		authToken = token
		fmt.Printf("✅ Auth token retrieved from Redis: %s...\n", token[:20])
	}
}

func TestUserRegistration(t *testing.T) {
	payload := map[string]interface{}{
		"username": "admin",
		"email":    "test_" + "@example.com",
		"password": "admin123",
		"role":     "developer",
	}

	resp, err := makeRequest("POST", "/users/register", payload, false, "")
	printResponse("User Registration", resp, err)
	defer resp.Body.Close()
}

func TestUserLogin(t *testing.T) {
	payload := map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	}

	resp, err := makeRequest("POST", "/users/login", payload, false, "")
	printResponse("User Login", resp, err)
	defer resp.Body.Close()
}

func TestCreateContainer(t *testing.T) {
	payload := map[string]interface{}{
		"container_name": "test-create-container",
		"image_name":     "nginx:stable-alpine-perl",
	}

	resp, err := makeRequest("POST", "/containers/create", payload, true, "")
	printResponse("Create Container", resp, err)
	defer resp.Body.Close()
}

func TestViewContainers(t *testing.T) {
	resp, err := makeRequest("GET", "/containers/view?field=container_id&order=desc", nil, true, "")
	printResponse("View Containers", resp, err)
	defer resp.Body.Close()

	var result struct {
		Data  interface{} `json:"data"`
		Total int         `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("Decode failed: %s", err)
		return
	}
	fmt.Println("Status Code", resp.StatusCode)
	fmt.Println("Total Containers:", result.Total)
	fmt.Println("Containers:", result.Data)
}

func TestUpdateContainer(t *testing.T) {
	payload := map[string]interface{}{
		"status": "OFF",
	}

	resp, err := makeRequest("PUT", "/containers/update/", payload, true, "")
	printResponse("Update Container", resp, err)
	defer resp.Body.Close()
}

func TestDeleteContainer(t *testing.T) {
	resp, err := makeRequest("DELETE", "/containers/delete/", nil, true, "")
	printResponse("Delete Container", resp, err)
	defer resp.Body.Close()
}

func TestExportContainers(t *testing.T) {
	resp, err := makeRequest("GET", "/containers/export?field=container_id&order=desc", nil, true, "")
	printResponse("Export Containers", resp, err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %s", err)
		return
	}

	fileName := "containers_exported-" + time.Now().Format("2006-01-02_15-04-05") + ".xlsx"

	if err := os.WriteFile(fileName, data, 0644); err != nil {
		log.Fatalf("Failed to save file: %s", err)
		return
	}
	fmt.Println("Excel file saved as", fileName)
}

func TestImportContainers(t *testing.T) {
	file, err := os.Open("test.xlsx")
	if err != nil {
		log.Fatalf("Open file failed: %s", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	formFile, err := writer.CreateFormFile("file", "test.xlsx")

	if err != nil {
		log.Fatalf("Create form file failed: %s", err)
		return
	}

	_, err = io.Copy(formFile, file)
	if err != nil {
		log.Fatalf("Copy file error: %s", err)
	}
	writer.Close()

	contentType := writer.FormDataContentType()
	resp, err := makeRequest("POST", "/containers/import", &buf, true, contentType)
	printResponse("Import Container", resp, err)
	defer resp.Body.Close()

	var result struct {
		SuccessCount      int      `json:"success_count"`
		SuccessContainers []string `json:"success_containers"`
		FailedCount       int      `json:"failed_count"`
		FailedContainers  []string `json:"failed_containers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Decode failed:", err)
		return
	}
	fmt.Println("Status Code", resp.StatusCode)
	fmt.Println("Success Count", result.SuccessCount)
	fmt.Println("Success Containers", result.SuccessContainers)
	fmt.Println("Failed Count", result.FailedCount)
	fmt.Println("Failed Containers", result.FailedContainers)
}

func TestEmailReport(t *testing.T) {
	resp, err := makeRequest("GET", "/report/mail?email=hung29032004@gmail.com&start_time=2025-01-01T00:00:00Z&end_time=2025-12-31T23:59:59Z", nil, true, "")
	printResponse("Email Report", resp, err)
	defer resp.Body.Close()
}

func TestUpdateUserPassword(t *testing.T) {
	payload := map[string]interface{}{
		"password": "newpassword123",
	}

	resp, err := makeRequest("PUT", "/users/update/password/user-id", payload, true, "")
	printResponse("Update User Password", resp, err)
	defer resp.Body.Close()
}

func TestUpdateUserRole(t *testing.T) {
	payload := map[string]interface{}{
		"role": "admin",
	}

	resp, err := makeRequest("PUT", "/users/update/role/user-id", payload, true, "")
	printResponse("Update User Role", resp, err)
	defer resp.Body.Close()
}

func TestUpdateUserScope(t *testing.T) {
	payload := map[string]interface{}{
		"isAdded": true,
		"scope":   []string{"container:view", "container:create"},
	}

	resp, err := makeRequest("PUT", "/users/update/scope/user-id", payload, true, "")
	printResponse("Update User Scope", resp, err)
	defer resp.Body.Close()
}

func TestDeleteUser(t *testing.T) {
	resp, err := makeRequest("DELETE", "/users/delete/user-id", nil, true, "")
	printResponse("Delete User", resp, err)
	defer resp.Body.Close()
}

func TestMain(t *testing.T) {
	// t.Run("UserRegistration", func(t *testing.T) {
	// 	TestUserRegistration(t)
	// 	TestSetupAuthToken(t)
	// })

	t.Run("UserLogin", func(t *testing.T) {
		TestUserLogin(t)
		TestSetupAuthToken(t)
	})

	// t.Run("CreateContainer", func(t *testing.T) {
	// 	TestCreateContainer(t)
	// })

	t.Run("ViewContainers", func(t *testing.T) {
		TestViewContainers(t)
	})

	// t.Run("UpdateContainer", func(t *testing.T) {
	// 	TestUpdateContainer(t)
	// })

	t.Run("ImportContainers", func(t *testing.T) {
		TestImportContainers(t)
	})

	t.Run("ExportContainers", func(t *testing.T) {
		TestExportContainers(t)
	})

	t.Run("EmailReport", func(t *testing.T) {
		TestEmailReport(t)
	})

	// t.Run("UpdateUserPassword", func(t *testing.T) {
	// 	TestUpdateUserPassword(t)
	// })

	// t.Run("UpdateUserRole", func(t *testing.T) {
	// 	TestUpdateUserRole(t)
	// })

	// t.Run("UpdateUserScope", func(t *testing.T) {
	// 	TestUpdateUserScope(t)
	// })

	// t.Run("DeleteUser", func(t *testing.T) {
	// 	TestDeleteUser(t)
	// })
	// t.Run("DeleteContainer", func(t *testing.T) {
	// 	TestDeleteContainer(t)
	// })
}
