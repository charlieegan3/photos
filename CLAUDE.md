# CLAUDE.md

## Overview

This is a personal photo sharing web application written in Go, designed as a spiritual successor to Instagram backup. It provides comprehensive photo organization with metadata extraction, local file storage, and features like trips, tags, devices, and location tracking.

## Development Commands

### Running the Development Server

```bash
make dev         # Starts server with hot reload on port 3000
```

### Testing

```bash
make test        # Run all tests
make test_watch  # Run tests in watch mode
make test_db     # Start local PostgreSQL test database in Docker
```

### Code Quality

```bash
make lint        # Run Go linter (golangci-lint)
make fmt         # Format code using goimports, gofumpt, dprint, and treefmt
```

### Database Operations

```bash
make new_migration MIGRATION_NAME=<name>  # Create new migration
make import                               # Import data
```

### Local Development

```bash
make local_bucket  # Start local bucket server for file storage
```

## Architecture

### Key Components

1. **Storage Layer** (`/bucket/`)
   - Uses gocloud.dev for local filesystem storage
   - Structure: `media/` (originals), `thumbs/` (generated), `device_icons/`, `lens_icons/`, `location_maps/`

2. **Database**
   - PostgreSQL with PostGIS extensions
   - Migrations in `/internal/pkg/database/migrations/`
   - Generic repository pattern in `/internal/pkg/database/` with type-safe CRUD operations
   - Models in `/internal/pkg/models/` (devices, lenses, media, posts, tags, trips, locations)

3. **Web Server** (`/internal/pkg/server/`)
   - Admin handlers: `/internal/pkg/server/handlers/admin/`
   - Public handlers: `/internal/pkg/server/handlers/public/`
   - Templates use Plush templating with layouts in `/internal/pkg/server/templating/`

4. **Media Processing**
   - EXIF extraction: `/internal/pkg/mediametadata/`
   - Image proxy/resizing: `/internal/pkg/imageproxy/`
   - Thumbnail generation happens automatically on upload

5. **External Integrations**
   - Geoapify for location maps: `/internal/pkg/geoapify/`
   - OAuth authentication: uses `charlieegan3/oauth-middleware`

### Important Patterns

- Generic repository pattern for database access (see `BaseRepository` in `/internal/pkg/database/repository.go`)
- Dependency injection through server initialization
- Command pattern using Cobra for CLI (`/cmd/`)
- Toolbelt integration for development tools (`/pkg/tool/`)

### Code Style

#### Error Handling

Prefer explicit error handling over inline error checks for better readability:

Instead of:

```go
if err := doSomething(); err != nil {
    return err
}
```

Prefer:

```go
err := doSomething()
if err != nil {
    return err
}
```

#### Comments

Comments should begin with the name of the thing being described and end in a period.

### Configuration

- Environment-specific configs: `config.dev.yaml`, `config.prod.yaml`, `config.test.yaml`
- Uses Viper for configuration management
- OAuth is configured via environment variables

#### Database Connection Types

The application supports both TCP and Unix socket connections to PostgreSQL:

**TCP Connection (default):**

```yaml
database:
  connectionString: postgres://localhost:5432/mydb
  params:
    dbname: mydb
    sslmode: disable
```

**Unix Socket Connection:**

```yaml
database:
  connectionString: postgres:///mydb  # Note: no host/port
  params:
    dbname: mydb
    host: /var/run/postgresql  # Unix socket directory
    sslmode: disable
```

### Testing Approach

- Test files follow `*_test.go` convention
- Test suites for major components (`*_suite.go`)
- Uses testify and go-testdeep for assertions
- Database tests require Docker (use `make test_db`)

## Development Tips

1. When modifying database schema, always create a migration using `make new_migration`
2. The admin interface is the primary way to manage content - accessible at `/admin`
3. Thumbnail generation is automatic but can be resource-intensive for large uploads
4. Activity data (GPX/TCX/FIT) parsing is in `/internal/pkg/activity/`
5. Static assets are embedded in the binary - see `/internal/pkg/server/static/`

## Templating System

The application uses a dual-layer templating system combining **Plush** and Go's `html/template`:

### Template Organization

- **Base templates**: `/internal/pkg/server/templating/` (`base.html`, `base.admin.html`)
- **Handler templates**: `/internal/pkg/server/handlers/{public|admin}/{module}/templates/`
- **File extension**: `.html.plush` for content templates

### Rendering Process

1. Templates embedded at compile time using `//go:embed`
2. Content rendered with Plush engine and context data
3. Result wrapped through template chain (e.g., content → admin → base)
4. Base templates use `{{.Body}}` and `{{.HeadContent}}` placeholders

### Helper Functions Available in Templates

- `to_string(arg)` - Convert any value to string
- `markdown(md)` - Convert markdown to HTML
- `truncate(s, length, ellipsis)` - Truncate strings
- `display_offset(media)` - Calculate CSS object-position for media based on dimensions
- `days_diff(t1, t2)` - Calculate days between dates
- `raw()` - Output unescaped HTML

### Plush Template Syntax

- `<%= variable %>` - Output with HTML escaping
- `<% code %>` - Code blocks for logic
- `<%= for (item) in items { %>` - Loops
- `<%= if (condition) { %>` - Conditionals
- `<% let variable = value %>` - Variable assignment

### Static Assets

- CSS bundle served at `/styles.css` (normalize + tachyons + custom styles)
- All static files embedded and served with ETag headers
- Responsive images use `<picture>` elements with multiple `srcset`
