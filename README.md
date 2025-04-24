# Gohost

This project is a backend API built with Go, created primarily as a learning exercise to explore various concepts in backend development. It simulates some features of a simple social media platform.

## Purpose

The main goal of this project was to practice and implement common backend patterns and technologies using the Go programming language.

## Features Implemented

*   **RESTful API:** Built using Go's standard `net/http` package.
*   **User Authentication:**
    *   User registration (`/api/users`) with email and password.
    *   Secure password hashing using `bcrypt`.
    *   User login (`/api/login`) providing JWT access and refresh tokens.
    *   User profile updates (`/api/users`).
    *   JWT validation middleware for protected routes.
    *   Token refresh (`/api/refresh`) and revocation (`/api/revoke`) mechanisms.
*   **"Gobits" (Posts) CRUD:**
    *   Create, Read (all, by author, by ID), and Delete operations for posts (`/api/gobits`).
    *   Authorization checks to ensure users can only delete their own gobits.
    *   Sorting capabilities for retrieving gobits.
*   **Database Interaction:**
    *   Integration with a PostgreSQL database using `database/sql` and the `github.com/lib/pq` driver.
    *   Type-safe database query generation using `sqlc`.
    *   Database schema migrations managed (following `goose` conventions).
*   **Webhook Handling:**
    *   An endpoint (`/api/strip/webhooks`) to receive and process external webhooks (e.g., for user upgrades).
    *   API Key authentication for securing the webhook endpoint.
*   **Input Validation:**
    *   Basic validation for request formats (JSON) and required fields.
    *   A specific endpoint (`/api/validate`) for text length checks and simple content moderation (bad word filtering).
*   **Configuration:**
    *   Environment variable management using `github.com/joho/godotenv`.

## Technologies Used

*   **Language:** Go
*   **Database:** PostgreSQL
*   **Core Libraries:**
    *   `net/http` (Web Server)
    *   `database/sql`, `github.com/lib/pq` (Database Access)
    *   `golang.org/x/crypto/bcrypt` (Password Hashing)
    *   `github.com/golang-jwt/jwt/v5` (JWT Handling)
    *   `github.com/google/uuid` (UUID Generation)
*   **Tooling:**
    *   `sqlc` (SQL to Go Code Generation)
    *   `goose` (Implied for Migrations)
    *   `godotenv` (Environment Variables)

## Key Learnings

This project provided hands-on experience with:

*   Building REST APIs from scratch in Go.
*   Implementing robust authentication flows with JWT (access/refresh tokens).
*   Securing user credentials with hashing.
*   Interacting with SQL databases and managing connections.
*   Leveraging tools like `sqlc` to improve database code safety and development speed.
*   Understanding and managing database schema changes (migrations).
*   Designing and securing webhook endpoints.
*   Structuring a Go web application with clear separation of concerns (handlers, auth logic, database logic).
*   Using middleware for cross-cutting concerns like metrics and authentication.
*   Managing application configuration securely.