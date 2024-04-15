# Radicle GitHub Actions Adapter

An adapter for the Radicle CI Broker (`rad:zwTxygwuz5LDGBq255RA2CbNGrz8`) that will report back to the broker the 
results from GitHub Actions that are executed for each push.

## Documentation

### Repository Setup

The steps for setting up a repository that works with the Radicle GitHub Actions Adapter can be found at 
[docs/project_setup.md](docs/project_setup.md).

## Getting Started

### Requirements

The minimum requirement for compiling and running this application are:
- Go v1.21
- make

Alongside the Radicle CI Broker, a Radicle node must be running on the same host. `radicle-httpd` is also required 
alongside the `RAD_SESSION_TOKEN` in order to add patch comments when a workflow completes.

### Configuration

The application uses configuration through Environment Variables. Here is a list with the details and the default
value for each one of them:

| EnvVar                        | Description                                                                  | Default Value           |
|-------------------------------|------------------------------------------------------------------------------|-------------------------|
| `LOG_LEVEL`                   | Set the log level of the application.<br>(`debug`, `info`, `warn`, `error`). | "info"                  |
| `RAD_HOME`                    | Path for radicle home directory.                                             | "~/.radicle"            |
| `RAD_HTTPD_URL`               | Public URL of radicle's HTTPD.                                               | "http://127.0.0.1:8080" |
| `RAD_SESSION_TOKEN`           | Session token for accessing Radicle API.                                     | ""                      |
| `GITHUB_PAT`                  | Personal access token for GitHub.                                            | ""                      |
| `WORKFLOWS_START_LAG_SECS`    | Lag time before giving up checking for GitHub's commit and workflows.        | 60                      |
| `WORKFLOWS_POLL_TIMEOUT_SECS` | Polling timeout for workflows completion.                                    | 1800                    |

`GITHUB_PAT` is not strictly required for public GitHub Repos.
For accessing **private repos** it should have at least read access for the
repo (`repo` access) and the actions/workflows (`workflow` access).
For accessing **public repos** it is highly advised to provide a GitHub Personal
Access Token (even without any permission) as GitHub has a
[rate limiting policy](https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api)
for accessing its API without any token.

`WORKFLOWS_START_LAG_SECS` is a necessary lag time as it is possible to push first to the radicle forge and then to
GitHub. This would generate an error as the adapter won't be able to find the commit at GitHub or the workflows
might have not been spawned at that time. So, the adapter will wait up to that time before checking at GitHub for 
the repo, the commit and the workflows.
 
### Running the application

In order to build the **Radicle GitHub Actions Adapter** use the provided makefile under project's root directory:

```bash
make build
```
This will generate the executable `/tmp/bin/radicle-github-actions-adapter`.

In order to build and run the application execute under project's root directory:

```bash
make run
```

Standard I/O is used for communication with the broker. Logging is directed to stderr.

> Radicle broker requires an executable of the adapter. Use `make build` to generate the binary.

### Application arguments

Application binary accepts specific arguments at init time. There are:

| Argument   | Example                                           | Description                                                                                                       | Default Value |
|------------|---------------------------------------------------|-------------------------------------------------------------------------------------------------------------------|---------------|
| `version`  | ./radicle-github-actions-adapter --version        | Prints only the binary's version and exits                                                                        | _empty_       |
| `loglevel` | ./radicle-github-actions-adapter --loglevel debug | Set the log level of the application.<br>(`debug`, `info`, `warn`, `error`)<br/>Overrides the Env Var `LOG_LEVEL` | "info"        |

### Versioning

Application uses SemVer version releases withVersion Control System's metadata. In order to specify a binary's version
it is generated from the revision of the source code and optionally the dirty flag which indicates if the binary
contains uncommitted changes. Here is an example of the version output: 

```
version: development, build_time: Fri Feb 16 16:53:24 EET 2024, revision: e63d3e19138f7165d11a5d046a1703ba06a69b23-dirty
```

A `development` version indicates that the specific build didn't produce from a specific released version.
Builds that originate from specific released versions contain information like this:

```
version: v0.5.1, build_time: Fri Feb 16 16:53:24 EET 2024, revision: e63d3e19138f7165d11a5d046a1703ba06a69b23
```

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
   "result": "<success|failure>"
}
```

If at least on job fails the result will be considered as failed.
In case of an unexpected a failure response will be replied back to the broker.

### Broker Message Protocol 

Adapter currently supports Radicle CI Broker message protocol versions:
> 1

## Contribute

Open an issue for discussing any issue or bug.
Open a patch for adding a new feature or any kind of fix.  
Use appropriate commands from makefile to ensure application correctness.  
For a complete list of makefile commands run:
```
make help
```
