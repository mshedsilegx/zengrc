# ZenGRC Attachment Downloader

This program downloads attachments and all associated metadata from the ZenGRC API.

## Usage

```bash
go run main.go -api-url <your-zengrc-api-url> -token <your-auth-token> -output-dir <output-directory>
```

### Flags

- `-api-url`: The URL of your ZenGRC API instance.
- `-token`: Your ZenGRC API authentication token.
- `-output-dir`: The directory where the attachments and metadata will be saved.