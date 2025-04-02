package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/keploy/keploy-review-agent/internal/analyzer"
	"github.com/keploy/keploy-review-agent/internal/api"
	"github.com/keploy/keploy-review-agent/internal/config"
	"github.com/keploy/keploy-review-agent/pkg/github"
)

type PullRequest struct {
	Number int `json:"number"`
	Head   struct {
		Sha string `json:"sha"`
	} `json:"head"`
	Base struct {
		Sha string `json:"sha"`
	} `json:"base"`
}

type Owner struct {
	Login string `json:"login"`
}

type Repository struct {
	Name  string `json:"name"`
	Owner Owner  `json:"owner"`
}

type Payload struct {
	Action     string      `json:"action"`
	PullRequest PullRequest `json:"pull_request"`
	Repository Repository  `json:"repository"`
}

// Extract owner and repo from GitHub pull request URL
func extractOwnerAndRepo(url string) (owner, repo string, err error) {
	// Regular expression to match the URL and capture owner and repo
	re := regexp.MustCompile(`https://api.github.com/repos/([^/]+)/([^/]+)/pulls/(\d+)`)
	matches := re.FindStringSubmatch(url)

	// If no matches, return an error
	if len(matches) < 3 {
		return "", "", fmt.Errorf("could not extract owner and repo from the URL")
	}

	// Return the owner and repo
	return matches[1], matches[2], nil
}

// Start HTTP server for handling GitHub events
func startServer(wg *sync.WaitGroup) {
	defer wg.Done()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome")
	})

	http.HandleFunc("/github", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Github-Event") == "pull_request" {
			// Read the request body
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Unable to read request body", http.StatusBadRequest)
				return
			}

			// Parse JSON into a map
			var data map[string]interface{}
			err = json.Unmarshal(body, &data)
			if err != nil {
				http.Error(w, "Unable to parse JSON", http.StatusInternalServerError)
				return
			}

			// Extract pull request action
			if action, ok := data["action"].(string); ok {
				if action == "opened" {
					// Get pull request number
					pullnumber := data["number"].(float64)
					fmt.Println(pullnumber)

					// Extract URL and owner/repo from the pull request data
					url := data["pull_request"].(map[string]interface{})["url"].(string)
					owner, repo, err := extractOwnerAndRepo(url)
					fmt.Printf("Owner: %s, Repo: %s\n", owner, repo)
					if err != nil {
						http.Error(w, "Failed to extract owner/repo", http.StatusInternalServerError)
						return
					}

					// Prepare payload for sending to localhost
					body := Payload{
						Action: "opened",
						PullRequest: PullRequest{
							Number: int(pullnumber),
							Head: struct {
								Sha string `json:"sha"`
							}{
								Sha: "abc123",
							},
							Base: struct {
								Sha string `json:"sha"`
							}{
								Sha: "def456",
							},
						},
						Repository: Repository{
							Name: repo,
							Owner: Owner{
								Login: owner,
							},
						},
					}

					// Marshal body to JSON
					jsonBody, err := json.Marshal(body)
					if err != nil {
						http.Error(w, "Failed to marshal body", http.StatusInternalServerError)
						return
					}
					fmt.Println(string(jsonBody))

					// Send POST request to localhost (via curl)
					go func() {
						// Run the curl command asynchronously to avoid blocking
						curlCmd := exec.Command("curl", "-X", "POST", 
							"-H", "Content-Type: application/json", 
							"-H", "X-GitHub-Event: pull_request", 
							"-H", "X-Hub-Signature-256: sha256=dummy", 
							"-d", string(jsonBody), 
							"http://localhost:8080/webhook/github")

						// Run the curl command
						output, err := curlCmd.CombinedOutput()
						fmt.Printf("Output: %s\n", output)
						if err != nil {
							fmt.Println("Error running curl command:", err)
							return
						}
					}()
				}
			}

			// Handle pull request number
			if pullnumber, ok := data["number"].(float64); ok {
				analyzer.PullRequestNumber(int(pullnumber))
				github.PullRequestNumber(int(pullnumber))
				fmt.Println(pullnumber)
			}
		}

		w.Write([]byte("success"))
	})

	// Start the server
	log.Printf("Server is running on port 6969")
	err := http.ListenAndServe(":6969", nil)
	if err != nil {
		log.Fatalln("Error starting server: ", err)
	}
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// WaitGroup to synchronize the servers and curl request
	var wg sync.WaitGroup

	// Start server for handling GitHub webhook events
	wg.Add(1)
	go startServer(&wg)

	// Setup router for the main server
	router := api.NewRouter(cfg)

	// Create HTTP server for the main application
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the main server
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Wait for the GitHub server to finish its processing
	wg.Wait()
	log.Println("Server exited properly")
}
