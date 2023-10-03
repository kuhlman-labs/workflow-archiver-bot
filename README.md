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

   ```sh
   git clone https://github.com/kuhlman-labs/workflow-archiver-bot.git
   ```

2. Navigate to the project directory:

   ```sh
   cd workflow-logs-archiver-bot
   ```

3. Build the Go executable:

   ```sh
   go build workflow-archiver.go
   ```

4. Set up a new GitHub App:
   - Go to the [GitHub Developer Settings](https://github.com/settings/apps) page.
   - Click on "New GitHub App".
   - Provide an appropriate name and description for your app.
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

   - Copy `setenv.sh.orig` to `setenv.sh` and fill out the values for authentication.
     - `AZURE_TENANT_ID`: The Azure tenant ID.
     - `AZURE_CLIENT_ID`: The Azure client ID.
     - `AZURE_CLIENT_SECRET`: The Azure client secret.
   - Run `. ./setenv.sh` to set the environment variables.
   - The [Service Principal](https://learn.microsoft.com/en-us/azure/role-based-access-control/role-assignments-portal?tabs=delegate-condition) needs the following permissions:
     - `Storage Blob Data Contributor` for the storage account.

> Note. [How to Create a Service Principal](https://learn.microsoft.com/en-us/purview/create-service-principal-azure)

7. Start the app:

   ```bash
   ./workflow-archiver
   ```

8. Configure the GitHub App:

   - Go back to the GitHub Developer Settings page.
   - Add a new webhook URL with the following details:
     - Payload URL: `http://<your-app-host>:<port>/api/github/hook`
     - Secret: the value of `WEBHOOK_SECRET` you set in `config.yml`
     - Select the individual events: `Workflow Run`.
   - Under Permissions, select the following:
     - `Read` for `Actions`.
   - Install the app on the orgs/repositories where you want to archive workflow logs.

## Usage

Once the Workflow Logs Archiver is set up and running, it will automatically listen for completed workflow runs in the repositories where it is installed. When a workflow run completes, the app will retrieve the logs, compress them, and upload the compressed logs to the specified Azure blob account.

Logs will be organized based on the repository and workflow run information, ensuring easy access and management of archived logs.

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on the GitHub repository.

## License

The Workflow Logs Archiver is open-source software released under the [MIT License](LICENSE).
