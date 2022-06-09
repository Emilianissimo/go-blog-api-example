# Getting started (let's GO (**_**))

Simple pure GO web app.

This program using SQlite3. If you need more - rewrite SQL or connect something like ORM, for example: gorm. Don't forget to download driver.

Tables are auto-migrating via first run.

For Ubuntu 20.04 LTS Focal Fossa (focal):

```bash
    sudo snap install go --classic
```

Run code:

```bash
    go run main.go
```

Or build it:

```bash
    go build main.go
```

If you have troubles with go.mod (you haven't this file lmao and you haven't dependencies (ofc))
-

Init go.mod file:

```bash
    go mod init package_name
```

Install sql (SQlite3) and some othe packages:

```bash
    go mod download github.com/mattn/go-sqlite3
    go get github.com/gorilla/mux # for better router
```
