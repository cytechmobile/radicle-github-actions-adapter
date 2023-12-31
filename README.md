# Radicle GitHub Actions Adapter

An adapter for the Radicle Broker that report back to the broker the results from GitHub Actions than are executed 
for each push.

## Documentation

### Repository Setup

The steps for setting up a repository that automatically updates both GitHub and radicle for all changes can be 
found at [docs/multi_forge_project_setup.md](docs/multi_forge_project_setup.md).

## Getting Started

### Requirements

The minimum requirement for compiling and running this application are:
- Go v1.21
- make

### Configuration

The application uses configuration through Environment Variables. Here is a list with the details and the default
value for each one of them:

| EnvVar                        | Description                                 | Default Value           |
|-------------------------------|---------------------------------------------|-------------------------|
| `RAD_HOME`                    | Path for radicle home directory.            | "~/.radicle"            |
| `RADICLE_NODE_URL`            | URL of Radicle node.                        | "http://127.0.0.1:8080" |
| `RADICLE_SESSION_TOKEN`       | Session token for accessing Radicle API.    | ""                      |
| `GITHUB_PAT`                  | Personal access token for GitHub            | ""                      |
| `WORKFLOWS_POLL_TIMEOUT_SECS` | Timout for polling for workflows completion | 600                     |

> `GITHUB_PAT` should have at least read access for the repo and the actions.
 
### Running the application

In order to build the **Radicle GitHub Actions Adapter** execute the provided makefile under project's root directory:

```bash
make build
```
This will generate the executable `/tmp/bin/radicle-github-actions-adapter`.

In order to build and run the application execute under project's root directory:

```bash
make run
```

Standard I/O is used for communication with the broker. Logging is directed to stderr.

> Radicle broker requires an executable of the adapter. Use `make build` to get only the binary

### Versioning

Application uses Version Control System's metadata. In order to specify a binary's version it is generated from the
revision of the source code and optionally the dirty flag which indicates if the binary contains uncommitted changes.
Here is an example version: `eecd4b8a194a24674b0ec30e60ef8c150918b975-dirty`

### Application arguments

Application binary accepts specific arguments at init time. There are:

| Argument   | Example                                             | Description                                                                    | Default Value |
|------------|-----------------------------------------------------|--------------------------------------------------------------------------------|---------------|
| `version`  | ./radicle-github-actions-adapter --version          | Prints only the binary's version and exits                                     | _empty_       |
| `loglevel` | ./radicle-github-actions-adapter --loglevel debug   | Set the log level of the application.<br/> (`debug`, `info`, `warn`, `error`)  | info          |


### Adapter Input - Output

Radicle GitHub Actions Adapter accepts messages and responds to Radicle CI Broker through std IO. The following messages
are exchanges throughout the adapter's runtime:

1. Incoming _Push Event Request_ or _Patch Event Request_ message as described at
   `rad:zwTxygwuz5LDGBq255RA2CbNGrz8/tree/doc/architecture.md`

2. Outgoing response message with the job ID:

```json
{
    "response": "triggered",
    "run_id": "<RUN-UUID>"
}
```
3. Outgoing response message with the response result:

```json
{
   "response": "finished",
   "result": "<success|failure>",
   "result_details": [
      {
         "workflow_id": "7369195312",
         "workflow_name": "Lint source code",
         "workflow_result": "success"
      },
      {
         "workflow_id": "7369195314",
         "workflow_name": "Run unit tests",
         "workflow_result": "failure"
      }
   ]
}
```

If at least on job fails the result will be considered as failed.

In case of an unexpected error the following response message is provided

```json
{
  "response": "finished",
  "result": {
    "error": "details of the error occurred"
  }
}
```

## Contribute

Open an issue for discussing any issue or bug.
Open a patch for adding a new feature or any kind of fix.  
Use appropriate commands from makefile to ensure application correctness.  
For a complete list of makefile commands run:
```
make help
```
