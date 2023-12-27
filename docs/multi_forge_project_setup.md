# Multi forge project setup

The goal of this document is to describe the process of creating a project/repository that will update both Radicle 
and GitHub for any change.

## Project Setup

The repository/project must be setup in a way that each update on the forge should update **both** GitHub and
Radicle. This way source code will be hosted in Radicle's network but also GitHub Actions will run within the GitHub.

Radicle GitHub Actions adapter will inform the Radicle project for any GitHub Actions' status through the Radicle Ci 
Broker.

The process for adding both push servers is to update `git remotes`. This can be done using the following commands 
within the project's root directory:

```bash
git remote add both rad://<RAD_ID>
git remote set-url --push both rad://<RAD_ID>/<NODE_ID>
git remote set-url --add --push both < https://github.com/user/repo.git > OR git remote set-url --add --push both < git@github.com:user/repo.git >
```

Where:
* <RAD_ID>: the rad id of the project (e.g. `z3VYhzZ9Vw4nqceS7Ns5vQbo3mctL`)
* <NODE_ID> the ID of the current running node (e.g. `z6MkkpTPzcq1ybmjQyQpyre15JUeMvZY6toxoZVpLZ8YarsB`)
* < https://github.com/user/repo.git > OR < git@github.com:user/repo.git >: the git repo URL

This will end up to this output of the `git remote -v` command:

```
» git remote -v
both	rad://z3VYhzZ9Vw4nqceS7Ns5vQbo3mctL (fetch)
both	rad://z3VYhzZ9Vw4nqceS7Ns5vQbo3mctL/z6MkkpTPzcq1ybmjQyQpyre15JUeMvZY6toxoZVpLZ8YarsB (push)
both	https://github.com/user/repo.git (push)
origin	https://github.com/user/repo.git (fetch)
origin	https://github.com/user/repo.git (push)
rad	rad://z3VYhzZ9Vw4nqceS7Ns5vQbo3mctL (fetch)
rad	rad://z3VYhzZ9Vw4nqceS7Ns5vQbo3mctL/z6MkkpTPzcq1ybmjQyQpyre15JUeMvZY6toxoZVpLZ8YarsB (push)
```

If the project hs started as a radicle project and then added github remotes it will end up to this output of the `git 
remote -v` command:

```
» git remote -v
both	rad://z2z4TxRWgHjKZzBjjAPceX59A7aC5 (fetch)
both	rad://z2z4TxRWgHjKZzBjjAPceX59A7aC5/z6MkkpTPzcq1ybmjQyQpyre15JUeMvZY6toxoZVpLZ8YarsB (push)
both	https://github.com/user/repo.git (push)
rad	rad://z2z4TxRWgHjKZzBjjAPceX59A7aC5 (fetch)
rad	rad://z2z4TxRWgHjKZzBjjAPceX59A7aC5/z6MkkpTPzcq1ybmjQyQpyre15JUeMvZY6toxoZVpLZ8YarsB (push)
```

Now every push should be done to using `both` remote. So pushing a commit should be invoked as

```bash
git push both main
```