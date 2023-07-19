# Workflow Logs Archiver

## Description
The Workflow Logs Archiver is a GitHub App written in Go that listens for completed workflow runs in a GitHub repository. Once a workflow run completes, the app retrieves the logs, compresses them, and uploads the compressed logs to an Azure blob account. This helps to centralize and store workflow logs in an easily accessible and scalable manner.

## Features
- Automatic retrieval and archiving of workflow logs.
- Compression of logs to reduce storage requirements.
- Seamless integration with GitHub repositories.
- Upload of compressed logs to Azure blob storage.
- Configurable settings for repository and Azure blob account details.

## Installation

To install and use the Workflow Logs Archiver GitHub App, follow these steps:

1. Clone the repository to your local machine:
   ```
   git clone https://github.com/your-username/workflow-logs-archiver.git
   ```

2. Navigate to the project directory:
   ```
   cd workflow-logs-archiver
   ```

3. Build the Go executable:
   ```
   go build
   ```

4. Set up a new GitHub App:
   - Go to the [GitHub Developer Settings](https://github.com/settings/apps) page.
   - Click on "New GitHub App".
   - Provide an appropriate name and description for your app.
   - Set the "Homepage URL" and "User authorization callback URL" to your desired values.
   - Choose the repositories where you want to use the app.
   - Set the necessary permissions (e.g., `workflow` for accessing workflow runs).
   - Generate a new private key and save it securely.

5. Configure the app:
   - Fill in the required environment variables in the `config.yml` file:
     - `v3_api_url`: The URL for the GitHub API v3. (Default: `https://api.github.com`) 
     - `integration_id`: The ID of your GitHub App.
     - `private_key`: The private key.
     - `webhook_secret`: A secret token to validate incoming webhooks.
     - `storage_account_name`: The name of the Azure storage account.
     - `address`: The address where the app will listen for incoming requests. (Default: 127.0.0.1)
     - `port`: The port where the app will listen for incoming requests. (Default: 8080)

6. Authenticate with Azure: 
   #### With a Service Principal
   - Set the following environment variables on the server where the app will run:
     - `AZURE_TENANT_ID`: The Azure tenant ID.
     - `AZURE_CLIENT_ID`: The Azure client ID.
     - `AZURE_CLIENT_SECRET`: The Azure client secret.
   - The Service Principal needs the following permissions:
     - `Storage Blob Data Contributor` for the storage account.

7. Start the app:
   ```bash
   ./workflow-logs-archiver
   ```

8. Configure the GitHub App webhook:
   - Go back to the GitHub Developer Settings page.
   - Under your app's settings, click on "Webhooks".
   - Add a new webhook URL with the following details:
     - Payload URL: `http://<your-app-host>:<port>/api/github/hook`
     - Secret: the value of `WEBHOOK_SECRET` you set in `config.yml`
     - Select the individual events: `Workflow Run`.
   - Under Permissions, select the following:
     - `Read` for `Actions`.

## Usage

Once the Workflow Logs Archiver is set up and running, it will automatically listen for completed workflow runs in the repositories where it is installed. When a workflow run completes, the app will retrieve the logs, compress them, and upload the compressed logs to the specified Azure blob account.

Logs will be organized based on the repository and workflow run information, ensuring easy access and management of archived logs.

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on the GitHub repository.

## License

The Workflow Logs Archiver is open-source software released under the [MIT License](LICENSE).