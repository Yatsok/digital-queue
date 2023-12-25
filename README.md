# Digique (Digital Queue)

## Prerequisites

Before you begin, ensure you have met the following requirements:

- Go installed on your local machine
- PostgreSQL database running

## Installing

Follow these steps to get the development environment running:

1. Clone the repository:

   ```bash
   git clone https://github.com/Yatsok/digital-queue.git
   cd digital-queue
   ```

2. Set up the environment variables:

   Below is a table containing the necessary environment variables and their example values for configuring the application:

   | Variable               | Example Value  |
   | ---------------------- | -------------- |
   | PORT                   | 8000           |
   | APP_ENV                | local          |
   | DB_HOST                | localhost      |
   | DB_PORT                | 5432           |
   | DB_DATABASE            | digique        |
   | DB_USERNAME            | postgres       |
   | DB_PASSWORD            | my_secure_pass |
   | JWT_SECRET_KEY         | jwt_key        |
   | JWT_REFRESH_SECRET_KEY | refresh_key    |

3. Run the application:

   ```bash
   make run
   ```

   OR

   ```bash
   make watch
   ```

   This will build the application, generate templates and start the server.

## Makefile Commands

### Building the application

```bash
# Build the application
make build

# Run the application
make run
```

### Testing and Cleanup

```bash
# Run tests
make test

# Clean up binary
make clean
```

### Live Reload

```bash
# Live reload during development
make watch
```

### Tailwind CSS

```bash
# Build Tailwind styles
make css

# Live update Tailwind styles
make css-watch
```

### Database Migrations

```bash
# Apply migrations
make migration-up

# Rollback migrations
make migration-down
```

#### Note

Make sure to set up your environment variables in the `.env` file.

## Built With

- Go
- PostgreSQL
- GORM
- HTMX
- HTML/CSS/JavaScript for the front-end

## Contributing

Contributions are welcome! Fork the repository and submit a pull request for any improvements, bug fixes, or new features.

## License

This project is licensed under the [MIT License](LICENSE).
