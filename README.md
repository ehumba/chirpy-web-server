# chirpy-web-server

## Introduction
Welcome to **Chirpy**! This web server allows clients to register an account, post short messages ("chirps"), and view other users' chirps. This project was created as part of the [Boot.dev curriculum](https://www.boot.dev/courses/learn-http-servers-golang). 

## Setup
1. In order to run the server, you will need to have **Go** and **PostgreSQL** installed. 

To install Go and PostgreSQL, please follow the instructions from the links below:

- **Install Go**: https://go.dev/doc/install
- **Install PostgreSQL**: https://learn.microsoft.com/en-us/windows/wsl/tutorials/wsl-database#install-postgresql

2. Clone or install the program by running `go install github.com/ehumba/chirpy-web-server@latest`. 

3. Create a .env file in the project directory with the required environment variables:
```
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
JWT_SECRET=your_jwt_secret
POLKA_KEY=your_polka_key
```

4. Run the database migrations (using goose or your migration tool).

5. Start the server:
`go run .`

## API instructions
The following is a list of the most important API endpoints and how to use them:

### User management
- **POST /api/users**
Create a new account with an email and a password by sending a request in the following format:

```
{
  "email": "user@example.com",
  "password": "example_password"
}
```

- **PUT /api/users**
Update the user data with the same request format as for creating a new account.

- **POST /api/login**
Login with your password and email.

Request format:

```
{
    "password": "example_password",
    "email": "user@example.com"
}
```

### Chirps
- **POST /api/chirps**
Create a new chirp with a text (body) of 140 characters or less.

```
{
    "body": "Your message"
}
```

- **GET /api/chirps** 
View all chirps. 
Optional query parameters:

`author_id` – filter by author
`sort=asc|desc` – sort by creation date (default: asc)

Example: 
`GET /api/chirps?author_id=123&sort=desc`


- **GET /api/chirps/{chirpID}**
View a specified chirp by its ID.

- **DELETE /api/chirps/{chirpID}**
Delete a chirp with the provided ID. 