package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

const (
	baseURL = "http://localhost:8080"
	timeout = 30 * time.Second
)

var (
	httpClient = &http.Client{Timeout: timeout}
	authToken  string
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

func makeRequest(method, endpoint string, payload interface{}, useAuth bool) (*http.Response, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

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
	defer resp.Body.Close()

	fmt.Printf("Status: %s\n", resp.Status)

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response body: %v\n", err)
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

	resp, err := makeRequest("POST", "/users/register", payload, false)
	printResponse("User Registration", resp, err)
}

func TestUserLogin(t *testing.T) {
	payload := map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	}

	resp, err := makeRequest("POST", "/users/login", payload, false)
	printResponse("User Login", resp, err)
}

func TestCreateContainer(t *testing.T) {
	payload := map[string]interface{}{
		"containerId":   "test-container",
		"containerName": "Test Container API Example",
		"status":        "ON", // or "OFF"
		"ipv4":          "192.168.1.100",
	}

	resp, err := makeRequest("POST", "/containers/create", payload, true)
	printResponse("Create Container", resp, err)
}

func TestViewContainers(t *testing.T) {
	resp, err := makeRequest("GET", "/containers/view", nil, true)
	printResponse("View Containers", resp, err)
}

func TestUpdateContainer(t *testing.T) {
	payload := map[string]interface{}{
		"status": "OFF",
	}

	resp, err := makeRequest("PUT", "/containers/update/test-container", payload, true)
	printResponse("Update Container", resp, err)
}

func TestDeleteContainer(t *testing.T) {
	resp, err := makeRequest("DELETE", "/containers/delete/test-container", nil, true)
	printResponse("Delete Container", resp, err)
}

func TestExportContainers(t *testing.T) {
	resp, err := makeRequest("GET", "/containers/export", nil, true)
	printResponse("Export Containers", resp, err)
}

// func TestImportContainers(t *testing.T) {
// }

func TestEmailReport(t *testing.T) {
	resp, err := makeRequest("GET", "/report/mail?email=test@example.com&start_time=2025-01-01T00:00:00Z&end_time=2025-12-31T23:59:59Z", nil, true)
	printResponse("Email Report", resp, err)
}

func TestUpdateUserPassword(t *testing.T) {
	payload := map[string]interface{}{
		"password": "newpassword123",
	}

	resp, err := makeRequest("PUT", "/users/update/password/user-id", payload, true)
	printResponse("Update User Password", resp, err)
}

func TestUpdateUserRole(t *testing.T) {
	payload := map[string]interface{}{
		"role": "admin",
	}

	resp, err := makeRequest("PUT", "/users/update/role/user-id", payload, true)
	printResponse("Update User Role", resp, err)
}

func TestUpdateUserScope(t *testing.T) {
	payload := map[string]interface{}{
		"isAdded": true,
		"scope":   []string{"container:view", "container:create"},
	}

	resp, err := makeRequest("PUT", "/users/update/scope/user-id", payload, true)
	printResponse("Update User Scope", resp, err)
}

func TestDeleteUser(t *testing.T) {
	resp, err := makeRequest("DELETE", "/users/delete/user-id", nil, true)
	printResponse("Delete User", resp, err)
}

func TestMain(t *testing.T) {
	fmt.Println("🚀 VCS-SMS API Examples")
	fmt.Println("======================================")

	// t.Run("UserRegistration", func(t *testing.T) {
	// 	TestUserRegistration(t)
	// 	TestSetupAuthToken(t)
	// })

	t.Run("UserLogin", func(t *testing.T) {
		TestUserLogin(t)
		TestSetupAuthToken(t)
	})

	t.Run("CreateContainer", func(t *testing.T) {
		TestCreateContainer(t)
	})

	t.Run("ViewContainers", func(t *testing.T) {
		TestViewContainers(t)
	})

	t.Run("UpdateContainer", func(t *testing.T) {
		TestUpdateContainer(t)
	})

	t.Run("ExportContainers", func(t *testing.T) {
		TestExportContainers(t)
	})

	// t.Run("ImportContainers", func(t *testing.T) {
	// 	TestImportContainers(t)
	// })

	t.Run("EmailReport", func(t *testing.T) {
		TestEmailReport(t)
	})

	t.Run("UpdateUserPassword", func(t *testing.T) {
		TestUpdateUserPassword(t)
	})

	t.Run("UpdateUserRole", func(t *testing.T) {
		TestUpdateUserRole(t)
	})

	t.Run("UpdateUserScope", func(t *testing.T) {
		TestUpdateUserScope(t)
	})

	// t.Run("DeleteUser", func(t *testing.T) {
	// 	TestDeleteUser(t)
	// })
	t.Run("DeleteContainer", func(t *testing.T) {
		TestDeleteContainer(t)
	})

	fmt.Println("\n======================================")
	fmt.Println("✅ API Examples Complete!")
}
