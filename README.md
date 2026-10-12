# Evermos Mini Project - Go Backend API

RESTful e-commerce backend API built with Go, Fiber,
GORM, and MySQL.

Developed as the final project for the Rakamin Academy
x Evermos Backend Developer Virtual Internship.

## Overview

## Features

- Authentication and authorization
- User and store management
- Product and category management
- Address management
- Transaction processing
- Product snapshots and stock management
- File uploads
- Pagination and filtering
- Indonesian region API integration

## Technology Stack

## Architecture

## Project Structure

```
evermos-mini-api/
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   ├── domain/
│   │   ├── user.go
│   │   ├── store.go
│   │   ├── product.go
│   │   ├── category.go
│   │   ├── address.go
│   │   ├── transaction.go
│   │   ├── transaction_item.go
│   │   └── product_log.go
│   │
│   ├── application/
│   │   ├── auth/
│   │   │   ├── service.go
│   │   │   └── dto.go
│   │   ├── user/
│   │   │   ├── service.go
│   │   │   └── dto.go
│   │   ├── store/
│   │   │   ├── service.go
│   │   │   └── dto.go
│   │   ├── product/
│   │   │   ├── service.go
│   │   │   └── dto.go
│   │   ├── category/
│   │   │   ├── service.go
│   │   │   └── dto.go
│   │   ├── address/
│   │   │   ├── service.go
│   │   │   └── dto.go
│   │   └── transaction/
│   │       ├── service.go
│   │       └── dto.go
│   │
│   ├── repository/
│   │   └── mysql/
│   │       ├── user_repository.go
│   │       ├── store_repository.go
│   │       ├── product_repository.go
│   │       ├── category_repository.go
│   │       ├── address_repository.go
│   │       └── transaction_repository.go
│   │
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/
│   │       │   ├── auth_handler.go
│   │       │   ├── user_handler.go
│   │       │   ├── store_handler.go
│   │       │   ├── product_handler.go
│   │       │   ├── category_handler.go
│   │       │   ├── address_handler.go
│   │       │   ├── transaction_handler.go
│   │       │   └── region_handler.go
│   │       ├── middleware/
│   │       │   ├── jwt.go
│   │       │   └── admin.go
│   │       ├── routes/
│   │       │   └── routes.go
│   │       ├── response/
│   │       │   └── response.go
│   │       └── validator/
│   │           └── validator.go
│   │
│   ├── infrastructure/
│   │   ├── database/
│   │   │   ├── mysql.go
│   │   │   ├── migrate.go
│   │   │   └── seed.go
│   │   ├── security/
│   │   │   ├── jwt.go
│   │   │   └── password.go
│   │   ├── storage/
│   │   │   └── local.go
│   │   ├── region/
│   │   │   └── client.go
│   │   └── config/
│   │       └── config.go
│   │
│   └── platform/
│       └── errors/
│           └── errors.go
│
├── migrations/
├── docs/
│   ├── architecture.md
│   ├── database.md
│   └── api.md
├── postman/
│   └── evermos-mini-api.postman_collection.json
├── uploads/
│   └── .gitkeep
├── tests/
│   └── integration/
├── .env.example
├── .gitignore
├── .golangci.yml
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── LICENSE
```

## Database Schema

## Prerequisites

## Getting Started

### Clone the Repository

### Configure Environment Variables

### Create the Database

### Run Migrations

### Seed Initial Data

### Start the Server

## API Documentation

## Testing with Postman

## Security Considerations

## Design Decisions

## Known Limitations

## Acknowledgments
