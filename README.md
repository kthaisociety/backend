# kthaisociety/backend

Next-generation backend for KTH AI Society website.

## Requirements

- Go 1.24.x
- Docker

## Setup

Use Docker compose to run the PostgreSQL database locally.

```bash
docker compose up -d
```

Use Go to run the program.

```bash
go run cmd/api/main.go
```

## Git Workflow

This project adhers to [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for commit messages. This is to ensure that the project is easily maintainable and that the commit history is clean and easy to understand.

## Notice about license

This project is licensed under the MIT license. See the [LICENSE](LICENSE) file for more information.

This does **not** apply to logos, icons, and images used in this project. They are the property of KTH AI Society and are not licensed for public, commercial, or personal use. If you wish to use them, please contact us at [contact@kthais.com](mailto:contact@kthais.com).
