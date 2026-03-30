# Tesseract
### Go API Boilerplate

A robust, production-ready backend template designed for modern SPAs (Nuxt, Next.js). This boilerplate focusing on extreme security, performance, and clean code principles.

### Tech Stack

-   **Language**: Go (Golang) 1.22+
-   **Database**: PostgreSQL + SQLC (Type-safe SQL)
-   **Cache**: Redis (Cache-aside pattern)
-   **Auth**: PASETO (Platform-Agnostic Security Tokens)
-   **Migrations**: goose (or your preferred tool)

### Key Features & Strengths

#### Security First

- **Fully Stateless PASETO Authentication**: Uses PASETO (V4) instead of JWT to eliminate common header/algorithm vulnerabilities. Optimized Access & Refresh token flow.
- **Advanced CSRF Protection**: Combined `SameSite=Lax` cookie policy with custom header validation (`X-Requested-With`) and strict `Origin/Referer` checks.
- **Smart Session Management**: Track active sessions with metadata (IP, Browser, OS).
- **Multi-Device Logout**: Logic to "Logout from other devices" by invalidating specific refresh token families in the database.
- **Rate Limiting & Payload Protection**: Global body size limits (preventing "Terabyte" attacks) and strict timeouts for all network operations.

#### Performance & Caching

- **Transparent Redis Layer**: Custom "Repository Decorator" pattern. The Service layer doesn't know about Redis; the Repository handles cache-aside logic automatically.
- **Automatic Cache Invalidation**: Smart invalidation on `UPDATE/DELETE` operations to ensure data consistency without "stale" data issues.
- **ETag Middleware**: Built-in hashing of GET responses to support 304 Not Modified, saving massive bandwidth for content-heavy responses.

#### Architecture

- **Strict Clean Architecture**: Clear separation between `Handlers` (HTTP), `Services` (Business Logic), and `Repositories` (Data Access).

- **Interface-Driven Design**: Decoupled packages to prevent circular dependencies and enable easy Unit Testing with mocks.

- **Type-Safe SQL**: Thanks to SQLC, your Go code always stays in sync with your schema. No more `interface{}` or reflection-heavy ORM magic.

- **Graceful Shutdown**: Handles OS signals to close DB and Redis connections cleanly without losing data.

#### Additional Features & Enterprise Readiness
- **Integrated OAuth 2.0 Flow**: Seamless social authentication (Google, GitHub, etc.) with automated account linking and session creation.

- **Header-based Internationalization (i18n)**: Automated localization of error messages and system notifications based on the Accept-Language header. Built-in middleware to detect and propagate the user's preferred locale.

- **Non-destructive Data Management (Soft Deletes)**: Robust protection against accidental data loss. Users and albums utilize a deleted_at pattern, allowing for easy data recovery and maintaining relational integrity without permanent removal.

- **Granular Access Control (RBAC)**: Built-in support for multiple user roles (Admin, User).

- **Permission-based Middleware**: Flexible pre-route authorization layer that prevents unauthorized access to sensitive endpoints before the request even reaches the business logic.


## Authentication endpoints

**GET** `/auth/google/provider`
Redirects to Google OAuth page

**POST** `/auth/refresh`
Exchange refresh token to new refresh and access tokens

**POST** `/auth/logout`
Deletes token cookies and refresh token from database

**POST** `/auth/logout-others`
Deletes all other refresh tokens from database (logout from other devices)

## Authenticated user endpoints

**GET** `/me/info`
Authenticated user info

**GET** `/me/list`
Authenticated user list of albums

**GET** `/me/deleted`
Authenticated user list of deleted albums

## Data endpoints

**GET** `/health`
Simple health check

**GET** `/users/{user_slug}/info`
Get user info by user slug

**GET** `/users/{user_slug}/list`
Get user list of albums by user slug

**GET** `/users/{user_slug}/profile`
Get user info and list of albums by user slug

**GET** `/albums/{user_slug}/{album_slug}`
Get the album data from user slug and album slug

## Modify endpoints

**PUT** `/users/{user_id}`
Update user info

**DELETE** `/users/{user_id}`
Delete user

**POST** `/albums/{album_id}`
Create new album

**PUT** `/albums/{album_id}`
Update album

**DELETE** `/albums/{album_id}`
Delete album

**POST** `/albums/{album_id}/restore`
Restore deleted album

**DELETE** `/albums/{album_id}/purge`
Purge deleted album

## Admin endpoints

**POST** `/admin/users/{user_id}/restore`
Restore deleted user. Requires admin role

**DELETE** `/admin/users/{user_id}`
Purge user with all albums and tokens. Requires admin role

## Playground endpoints

For development purposes there is `/cmd/api/routed_playground.go` file with `Playground endpoints`. Should be deleted in production

**GET** `/playground/create_admin`
Creates `admin` user with 3 albums (public, private and shared for user) (`/users/admin/profile`)

**GET** `/playground/create_user`
Creates `user` user with 3 albums (public, private and shared for admin) (`/users/user/profile`)

**GET** `/playground/get_admin_cookies`
Get access and refresh tokens for admin

**GET** `/playground/get_user_cookies`
Get access and refresh tokens for user

**GET** `/playground/clear_cookies`
Deletes all token cookies

**GET** `/playground/clear_cache`
Flushes all redis cache data

## Installation

1. Clone the project:

```
git clone https://github.com/volkar/go-api.git
```

2. Go to the project's folder

```
cd go-api-bolierplate
```

3. Copy .env.example to .env

```
cp .env.example .env
```

4. Edit .env

```
Generate required random strings
Fill Postgres credentials
Fill Google OAuth credentials
```

5. Run database migration (with [goose](https://github.com/pressly/goose) for example)

```
goose up
```

6. Run your golang app (via [air](https://github.com/air-verse/air) for example):

```
air
```

7. Open address in curl or Yaak/Insomnium.

```
curl -X GET http://localhost:1337/health
```

8. Now you can use `Playground routes` to create `admin` and `user` and test this API.

## Contact me

You always welcome to write me

-   E-mail: sergey@volkar.ru
-   Telegram: @sergeyvolkar

All PR are welcome!