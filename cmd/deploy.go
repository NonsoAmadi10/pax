package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/briandowns/spinner"
	"github.com/digitalocean/godo"
	"github.com/spf13/cobra"
)

var buildCommand string
var deployCommand string
var appName string
var portNumber string

var deployCmd = &cobra.Command{
	Use:   "deploy [path]",
	Short: "Deploy your frontend to digital ocean",
	Args:  cobra.ExactArgs(1),
	Run:   runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

func runDeploy(cmd *cobra.Command, args []string) {
	path := args[0]

	token := os.Getenv("ACCESS_TOKEN")

	if token == "" {
		fmt.Println("Error: ACCESS_TOKEN is missing")
		return
	}

	// create a digital ocean client
	client := godo.NewFromToken(token)

	fmt.Println("What do you want to call this app?")
	fmt.Scanln(&appName)

	fmt.Println("Enter the build command:")
	fmt.Scanln(&buildCommand)

	fmt.Println("Building the project...")

	if err := executeCommand(buildCommand, path); err != nil {
		fmt.Println("Build failed:", err)
		return
	}
	fmt.Println("What port number will it run on ?")
	fmt.Scanln(&portNumber)

	port, err := strconv.Atoi(portNumber)
	if err != nil {
		fmt.Printf("invalid port number: %v", err)
	}
	// Prepare the App platform deployment request
	appReq := &godo.AppCreateRequest{
		Spec: &godo.AppSpec{
			Name: appName,
			Services: []*godo.AppServiceSpec{
				{
					Name:         appName,
					RunCommand:   deployCommand,
					SourceDir:    path,
					BuildCommand: buildCommand,
					HTTPPort:     int64(port),
				},
			},
		},
	}

	fmt.Println("Deploying to DigitalOcean...")

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)

	s.Start()

	ctx := context.TODO()

	app, _, err := client.Apps.Create(ctx, appReq)

	if err != nil {
		s.Stop()
		fmt.Println("Error deploying app:", err)
		return
	}

	s.Stop()

	fmt.Println("App is being deployed. Please hold on...")

	// Get deployment Status
	for {
		time.Sleep(10 * time.Second)
		updatedApp, _, err := client.Apps.Get(ctx, app.ID)

		if err != nil {
			fmt.Println("Error fetching app status:", err)
			return
		}

		status := updatedApp.LiveURL

		if status != "" {
			fmt.Printf("Deployment succesful! Your app is live here at %s\n", status)
			return
		}

		fmt.Println("Deployment in progress...")
	}
}

func executeCommand(cmd string, dir string) error {
	command := exec.Command("bash", "-c", cmd)
	command.Dir = dir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}
