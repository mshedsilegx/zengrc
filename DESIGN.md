# Specifications

Examine API specs at: https://docs.api.zengrc.com/api/v2/swagger-v3.json and design a program in Golang to download attachments and all associated metadata. Produce a folder per record ID, containing a `metadata.json` file for metadata information (audit history, date, etc), and all attachments associated with that specific record. Investigate the easiest method to authenticate.

The list of record IDs to process must be returned by an API call, and iterated through via goroutines Proceed carefully, with documentation validation at every step. The final product must be a Go binary that is fully functional. Run `go vet ./...` and correct any issues before committing. Implement a `-version` flag. It defaults to `dev`, and can be overwritten at build time by `-X main.version=` .

Search for the following documents to get knowledge: ZenGRC API authentication ZenGRC REST API documentation ZenGRC API request object attachments ZenGRC API request record metadata ZenGRC API request list of records

# Plan

1. Project Scaffolding

Create the basic structure for the Go application. This includes initializing a Go module, creating a `main.go` file to house the primary logic, and a `README.md` to provide clear instructions on how to use the program.

2. Configuration Management
The application will require user-specific inputs to function correctly. Implement a command-line interface using the `flag` package to accept the ZenGRC API URL, an authentication token, and the desired output directory for the downloaded files. This approach ensures that sensitive information is not hardcoded into the application.

3. API Client Implementation
Develop a dedicated API client to handle all interactions with the ZenGRC API. This client will manage HTTP requests and include Go structs that represent the expected JSON responses for Request and Attachment objects, based on the now-retrieved API documentation.

4. Record Retrieval and Processing
The core functionality of the application will be to fetch and process records. Implement a function to retrieve the initial list of Request objects from the API. To optimize performance, use a worker pool pattern with goroutines to process these records concurrently. 

For each record, the application will:
- Create a dedicated folder named after the record's ID.
- Fetch the full metadata for the record and save it as a `metadata.json` file within the folder.
- Retrieve the list of attachments associated with the record.
- Download each attachment into the record's folder.

5. Pre-commit steps
Complete pre commit steps to make sure proper testing, verifications, reviews and reflections are done.

6. Submission
Once the application is fully functional and tested, present the code with a descriptive commit message.

## Code review

Perform a deep analysis of the codebase, comparing it against the ZenGRC API specifications and evaluating it for security, performance, and modularity. Find below a set of recommendations for improvement.

### In-Depth Code Review

Overall, the application is well-structured and provides a solid foundation. The separation of concerns between `main.go` (application logic) and `client.go` (API interaction) is good practice. The use of a worker pool for concurrency is also an excellent choice for this type of task.

However, there are several areas for improvement to elevate the code to a production-grade standard.

#### 1. Security
- Directory Permissions: The code uses `os.ModePerm` (which translates to `0777`) when creating directories. This is overly permissive and could be a security risk in a multi-user environment. Permissions should be as restrictive as possible.
- Input Sanitization: While the use of `filepath.Join` helps prevent basic path traversal attacks on the output directory, no validation is performed on the `-api-url` flag. A malformed URL could cause unexpected behavior.
- File Overwriting: The application currently overwrites existing files without warning. If a record is processed twice, any existing metadata or attachments in its directory will be overwritten.

#### 2. Performance & Scalability
- HTTP Client Configuration: The current `http.Client` uses default transport settings. For a high-volume download application, this can be inefficient. The default transport doesn't optimize for connection reuse (keep-alives) as effectively as a customized one could, potentially leading to performance degradation under heavy load due to excessive connection setup/teardown.
- Concurrency Model: The worker pool is a great start. However, the main goroutine that fetches the list of requests does so sequentially, page by page. This means the workers may sit idle waiting for the next batch of requests to be discovered. This is a limitation of the API's pagination model, but the application's overall throughput is capped by this initial sequential discovery.

#### 3. Modularity & Maintainability
- API Compliance: The Go structs for Request and Attachment are currently minimal. The `metadata.json` file will therefore be incomplete, missing a lot of valuable information available from the API (e.g., `created_at`, `due_date`, `status`, `owners`, etc., for a Request). To be truly useful, these structs should fully reflect the schemas in the API documentation.
- Error Handling: The worker goroutines (`processRequest`) currently log errors to the console and then simply return. This is not ideal because:
    + The main application is unaware that failures are occurring.
    + There is no mechanism to retry failed requests or to report a summary of failures at the end.
    + A more robust solution would propagate these errors back to the main goroutine for centralized handling.
- Hardcoded API Paths: API endpoint paths (e.g., `/api/v2/requests`) are hardcoded as strings within the client methods. This makes them harder to manage and update. Consolidating them as constants would improve maintainability.

#### 4. Logic & Code Health
- Race Conditions/Deadlocks: The current concurrency model is sound and does not appear to have race conditions or deadlocks.
- Unused Code: The `strings` package was previously imported but not used. I have already corrected this. There are no other unused parameters or deprecated functions.

## Plan for Recommendations
Based on the review, I have designed the following plan to address these points, prioritized by impact and importance.

### High Priority (Correctness & Security)

1. Expand Data Models: I will update the Request and Attachment structs in `client.go` to include all relevant fields as defined in the `swagger-v3.json` API specification. This will ensure the `metadata.json` file is comprehensive and compliant.
2. Set Secure File Permissions: I will change the directory creation permission from `os.ModePerm` (`0777`) to the more secure `0755`.

### Medium Priority (Robustness & Modularity)

3. Implement Robust Error Handling: I will refactor the concurrency model to use an error channel. Workers will send any errors encountered during processing back to the main goroutine, which will collect and log them at the end, providing a clear summary of any failed records or downloads.
4. Centralize API Paths: I will define all API endpoint paths as constants at the top of the `client.go` file to improve code organization and ease of maintenance.

### Low Priority (Performance & Polish)
5. Add File Overwrite Protection: I will add a new boolean command-line flag, `-overwrite`. If this flag is false (the default), the application will check if an attachment already exists before downloading and will skip it to prevent overwriting.
6. Optimize HTTP Client: I will configure the `http.Client` with a custom `http.Transport` to optimize connection pooling and reuse (e.g., setting `MaxIdleConns` and `IdleConnTimeout`), improving performance for large numbers of API calls.

## Documentation review

Provide comprehensive documentation in `README.md` with sections: Application overview and objectives, architecture and design choices, command line arguments (list with description, type and defaults), examples on how to use. Make sure that the code itself is clearly documented inline, with clear explanation of the objective and description of core sections and functionality. Perform a documentation review to make sure it is comprehensive and accurate, by scanning and reconciling against the code base. Create a file `CHANGELOG.md` and document all changes implemented in this session, groupes under a common section representing the date of changes (yyyy-mm-dd)

Add a section in `README.md`, Metadata Details detailing all fields extracted from the metadata via the API, in a tabular format, with data type and all relevant information

Create a new section in `README.md`, called" Attachment Management, to explain how are the attachments associated with the metadata
