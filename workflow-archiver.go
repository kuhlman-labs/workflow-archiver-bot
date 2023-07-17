package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"compress/gzip"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/google/go-github/v53/github"
	"github.com/google/uuid"
	"github.com/gregjones/httpcache"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/rcrowley/go-metrics"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

type WorkflowHandler struct {
	githubapp.ClientCreator
	storageAccountName string
}

type Config struct {
	Server HTTPConfig       `yaml:"server"`
	Github githubapp.Config `yaml:"github"`
	Azure  Azure            `yaml:"azure"`
}

type Azure struct {
	storageAccountName string `yaml:"storage_account_name"`
}

type HTTPConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

func readConfig(path string) (*Config, error) {
	var c Config

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed reading server config file: %s", path)
	}

	if err := yaml.UnmarshalStrict(bytes, &c); err != nil {
		return nil, errors.Wrap(err, "failed parsing configuration file")
	}

	return &c, nil
}

func (h *WorkflowHandler) Handles() []string {
	return []string{"pull_request"}
}

func (h *WorkflowHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	// Parse the workflow run event
	var event github.WorkflowRunEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse workflow run event")
	}

	// Get the installation ID
	installationID := githubapp.GetInstallationIDFromEvent(&event)

	// Prepare the context
	ctx = githubapp.DefaultContextDeriver(ctx)

	// Check if workflow run is completed
	if *event.Action != "completed" {
		zerolog.Ctx(ctx).Info().Msgf("Workflow run %d is not completed", *event.WorkflowRun.ID)
		return nil
	}

	// Get the installation client
	client, err := h.NewInstallationClient(installationID)
	if err != nil {
		return err
	}

	// Get the workflow run
	workflowRun, _, err := client.Actions.GetWorkflowRunByID(ctx, *event.GetRepo().Owner.Login, *event.GetRepo().Name, *event.WorkflowRun.ID)
	if err != nil {
		return errors.Wrap(err, "failed to get workflow run")
	}

	// Check if workflow run is successful
	if *workflowRun.Conclusion != "success" {
		zerolog.Ctx(ctx).Info().Msgf("Workflow run %d is not successful", *event.WorkflowRun.ID)
	}

	// Get Log URL from workflow run
	logURL, _, err := client.Actions.GetWorkflowRunLogs(ctx, *event.GetRepo().Owner.Login, *event.GetRepo().Name, *event.WorkflowRun.ID, true)
	if err != nil {
		return errors.Wrap(err, "failed to get workflow run log URL")
	}

	// Get the log
	log, err := http.Get(logURL.String())
	if err != nil {
		return errors.Wrap(err, "failed to get workflow run logs")
	}

	blobURL := fmt.Sprintf("https://%s.blob.core.windows.net/", h.storageAccountName)

	// Log to Azure Blob Storage
	err = h.logToAzureBlobStorage(log, blobURL, *event.GetRepo().Name, *event.GetRepo().Owner.Login, *event.WorkflowRun.ID)
	if err != nil {
		return errors.Wrap(err, "failed to log to Azure Blob Storage")
	}

	return nil
}

// Log to Azure Blob Storage
func (h *WorkflowHandler) logToAzureBlobStorage(log *http.Response, blobURL, repoName, orgName string, workflowRunID int64) error {
	// Create a context object for the request
	ctx := context.Background()

	//convert http.Response to []byte
	body, err := io.ReadAll(log.Body)
	if err != nil {
		return errors.Wrap(err, "failed to convert http.Response to []byte")
	}

	//compress log
	compressedBody, err := compress(body)
	if err != nil {
		return errors.Wrap(err, "failed to compress []byte")
	}

	// Create a default credential object using the default Azure Identity
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return errors.Wrap(err, "failed to get Azure credential")
	}

	// Create a azure blob client
	client, err := azblob.NewClient(blobURL, credential, nil)
	if err != nil {
		return errors.Wrap(err, "failed to get Azure Blob client")
	}

	// Create the container
	containerName := fmt.Sprintf("%s-%s", orgName, repoName)
	fmt.Printf("Creating a container named %s\n", containerName)
	_, err = client.CreateContainer(ctx, containerName, nil)
	if bloberror.HasCode(err, bloberror.ContainerAlreadyExists) {
		fmt.Printf("A container named %s already exists.\n", containerName)
	} else if err != nil {
		return errors.Wrap(err, "failed to create container")
	}

	// Create a unique name for the blob
	blobName := fmt.Sprintf("%s-%s.log.gz", time.Now().Format("20060102150405"), uuid.New().String())

	// Upload to data to blob storage
	fmt.Printf("Uploading a blob named %s\n", blobName)
	_, err = client.UploadBuffer(ctx, containerName, blobName, compressedBody, &azblob.UploadBufferOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to upload buffer")
	}

	return nil
}

// Compress log with gzip
func compress(body []byte) ([]byte, error) {

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(body); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func main() {
	config, err := readConfig("config.yml")
	if err != nil {
		panic(err)
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &logger

	metricsRegistry := metrics.DefaultRegistry

	cc, err := githubapp.NewDefaultCachingClientCreator(
		config.Github,
		githubapp.WithClientUserAgent("workflow-archiver-bot/1.0.0"),
		githubapp.WithClientTimeout(3*time.Second),
		githubapp.WithClientCaching(false, func() httpcache.Cache { return httpcache.NewMemoryCache() }),
		githubapp.WithClientMiddleware(
			githubapp.ClientMetrics(metricsRegistry),
		),
	)
	if err != nil {
		panic(err)
	}

	workflowHandler := &WorkflowHandler{
		ClientCreator:      cc,
		storageAccountName: config.Azure.storageAccountName,
	}

	webhookHandler := githubapp.NewDefaultEventDispatcher(config.Github, workflowHandler)

	http.Handle(githubapp.DefaultWebhookRoute, webhookHandler)

	addr := fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port)
	logger.Info().Msgf("Starting server on %s...", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}
