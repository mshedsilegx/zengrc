# ZenGRC Attachment Downloader

## 1. Application Overview and Objectives

The ZenGRC Attachment Downloader is a command-line application written in Go that downloads attachments and all associated metadata from the ZenGRC API.

The primary objectives of this application are:
- To provide a reliable way to back up or archive ZenGRC records and their attachments.
- To process records concurrently to maximize download speed and efficiency.
- To offer a secure and configurable way to interact with the ZenGRC API.

For each record found, the application creates a dedicated folder, saves the complete record metadata as a `metadata.json` file, and downloads all associated attachments into that folder.

## 2. Architecture and Design Choices

The application is designed with a focus on performance, security, and maintainability.

- **Modularity:** The codebase is split into two main files:
    - `main.go`: Contains the application's entry point, command-line flag parsing, and the concurrency logic (worker pool).
    - `client.go`: Contains a dedicated API client for all interactions with the ZenGRC API, separating the application logic from the API communication logic.

- **Concurrency:** The application uses a worker pool pattern to process records concurrently. This allows for multiple records to be downloaded at the same time, significantly improving performance when dealing with a large number of records. Errors from concurrent workers are collected in a dedicated channel and reported at the end of the execution, ensuring that no failure goes unnoticed.

- **Security:**
    - **Secure File Permissions:** Directories are created with `0755` permissions, and files with `0644`, to prevent unauthorized access in a multi-user environment.
    - **No Hardcoded Credentials:** The API token is passed via a command-line flag, preventing sensitive information from being stored in the source code.
    - **File Overwrite Protection:** By default, the application will not overwrite existing files, preventing accidental data loss. This can be overridden with the `-overwrite` flag.

- **Performance:** The HTTP client is configured with a custom transport to optimize connection pooling and reuse, which is crucial for an application that makes a large number of API calls.

## 3. Command-Line Arguments

The application is configured using the following command-line flags:

| Flag          | Type    | Default                | Description                                                              |
|---------------|---------|------------------------|--------------------------------------------------------------------------|
| `-api-url`    | string  | (none)                 | **(Required)** The URL of your ZenGRC API instance (e.g., `https://acme.api.zengrc.com`). |
| `-token`      | string  | (none)                 | **(Required)** Your ZenGRC API authentication token in the format `key_id:key_secret`. |
| `-output-dir` | string  | `./zengrc_attachments` | The directory where the attachments and metadata will be saved.            |
| `-workers`    | int     | `5`                    | The number of concurrent workers to use for downloading.                 |
| `-overwrite`  | bool    | `false`                | If set to `true`, the application will overwrite existing files.         |

## 4. Examples

### Basic Usage

The following command will download all records and their attachments to the default `./zengrc_attachments` directory, using 5 concurrent workers.

```bash
go run main.go client.go \
  -api-url "https://your-instance.api.zengrc.com" \
  -token "your_key_id:your_key_secret"
```

### Custom Output Directory and Worker Count

This example downloads the records to a custom directory (`/path/to/my/backup`) and uses 10 concurrent workers.

```bash
go run main.go client.go \
  -api-url "https://your-instance.api.zengrc.com" \
  -token "your_key_id:your_key_secret" \
  -output-dir "/path/to/my/backup" \
  -workers 10
```

### Overwriting Existing Files

If you need to re-download all files and overwrite any that already exist, use the `-overwrite` flag.

```bash
go run main.go client.go \
  -api-url "https://your-instance.api.zengrc.com" \
  -token "your_key_id:your_key_secret" \
  -overwrite
```