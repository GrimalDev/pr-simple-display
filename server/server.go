package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	CreatedAt string `json:"createdAt"`
	Status    string `json:"state"`
}

// getPullRequests fetches open pull requests and their statuses
func getPullRequests() ([]PullRequest, error) {
	// Command to get open pull requests using GitHub CLI
	// Get the repository name from the .env file
	repo := os.Getenv("REPO")
	cmd := exec.Command("gh", "pr", "list", "--repo", repo, "--json", "number,title,createdAt,state")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	// Parse JSON output
	var pullRequests []PullRequest
	err = json.Unmarshal(output, &pullRequests)
	if err != nil {
		return nil, err
	}

	return pullRequests, nil
}

// formatPullRequestToHTML formats a pull request to the required HTML
func formatPullRequestToHTML(pr PullRequest) string {
	date := pr.CreatedAt[:10] + " " + pr.CreatedAt[11:19]
	return fmt.Sprintf("<li><div class=\"col col-1\">%d</div><div class=\"col col-2\">%s</div><div class=\"col col-3\">%s</div><div class=\"col col-4\">%s</div></li>", pr.Number, pr.Title, pr.Status, date)
}

// handleSSE handles Server-Sent Events (SSE) connections
func handleSSE(c echo.Context) error {
	// Set the necessary headers for SSE
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
	c.Response().Header().Set(echo.HeaderConnection, "keep-alive")
	c.Response().WriteHeader(http.StatusOK)

	// Create a ticker to send SSE events every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request().Context().Done():
			// Client disconnected
			return nil
		case <-ticker.C:
			// Fetch pull requests
			pullRequests, err := getPullRequests()
			if err != nil {
				return err
			}

			// Format pull requests to HTML and create SSE event data
			var formattedHTML string
			for _, pr := range pullRequests {
				formattedHTML += formatPullRequestToHTML(pr)
			}

			// Send SSE event with event name "pr"
			fmt.Fprintf(c.Response(), "event: pr\ndata: %s\n\n", formattedHTML)
			c.Response().Flush()
		}
	}
}

func main() {
	// Create a new Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	err := godotenv.Load(".env")
	if err != nil {
		e.Logger.Fatal("Error loading .env file")
	}

	// Enable CORS for all origins
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.Static("/public", "public")

	// SSE endpoint
	e.GET("/sse", handleSSE)

	e.Logger.Fatal(e.Start(":3434"))
}
