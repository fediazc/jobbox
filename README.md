# Jobbox

Jobbox is a job application tracker that makes it easy to stay organized while you're on the hunt. Ditch the messy spreadsheet! Make your application tracking easy, with Jobbox.

# How to run

**Prerequisites**:

* Go 1.21+
* MySQL 8.1.0+

## Set up the database

Run the following SQL commands (using MySQL) to create the database and the necessary tables:

```SQL
CREATE DATABASE cazar CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE cazar;

CREATE TABLE users (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    hashed_password CHAR(60) NOT NULL
);

ALTER TABLE users ADD CONSTRAINT users_uc_email UNIQUE (email);

CREATE TABLE jobs (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    user_id INTEGER NOT NULL,
    company VARCHAR(255) NOT NULL,
    job_role VARCHAR(255) NOT NULL,
    commute VARCHAR(255) NOT NULL,
    application_status VARCHAR(255) NOT NULL,
    location VARCHAR(255) NOT NULL,
    date_applied DATETIME NOT NULL,
    notes TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE sessions (
    token CHAR(43) PRIMARY KEY,
    data BLOB NOT NULL,
    expiry TIMESTAMP(6) NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);
```

## Set up TLS

The project is configured to use HTTPS, so you will need to create a self-signed TLS certificate in order to run it. If you are on Linux/MacOS, make sure you're in the base directory for the project and run the following commands on a terminal: 

```
mkdir tls
cd tls
go run /usr/local/go/src/crypto/tls/generate_cert.go --rsa-bits=2048 --host=localhost
```

If you are on windows or you installed Go in some other directory, replace `/usr/local/go/src/crypto/tls/generate_cert.go` with wherever your `generate_cert.go` is located.

## Running the project

Make sure you are in the base directory for the proejct and run the following command:

```
go run ./cmd/web
```
