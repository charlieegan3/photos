# photos

[photos.charlieegan3.com](https://photos.charlieegan3.com) is a personal, single user photo sharing site
which I built and host for myself. The application has a number of features I find valuable:

- [Photo Map](https://photos.charlieegan3.com/locations) - Interactive map view of all geotagged posts
- [Trips](https://photos.charlieegan3.com/posts/period) - Browse posts from specific trips and date ranges
- [Search](https://photos.charlieegan3.com/posts/search) - Find posts by tag, description, and location
- [On This Day](https://photos.charlieegan3.com/posts/on-this-day) - Discover posts from this day in previous years
- [Browse by Device](https://photos.charlieegan3.com/devices) - View posts by camera or device used
- [Browse by Lens](https://photos.charlieegan3.com/lenses) - Filter posts by specific lenses
- [Tags](https://photos.charlieegan3.com/tags) - Browse posts organized by tags
- [Random](https://photos.charlieegan3.com/random) - Discover a random post
- [RSS](https://photos.charlieegan3.com/rss.xml) - Subscribe to updates

The app is formed of a Go application. The project is the
spiritual successor of a project I built to back up my Instagram account
[in 2018](https://charlieegan3.com/posts/2018-03-04-backing-up-instagram).

## Config

Should you want to run an instance of this application yourself,
the required configuration file looks like this:

```yaml
environment: development # or production
admin:
  auth:
    permitted_email_suffix: "@example.com"
server:
  address: "localhost"
  port: "3000"
  https: false
geoapify:
  url: https://maps.geoapify.com/v1/staticmap
  key: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
database:
  migrationsTable: schema_migrations_photos
  connectionString: postgresql://postgres:password@localhost:5432
  params:
    dbname: cms_dev
    sslmode: disable
bucket:
  url: file://./bucket
notification_webhook:
  endpoint: https://example.com
```

### Authentication

The application supports two authentication modes based on the environment:

- **Development** (`environment: development`): No authentication required for admin routes (`/admin/*`)
- **Production** (`environment: production`): Admin routes require authentication via reverse proxy

In production mode, the reverse proxy must set an `X-Email` header containing the authenticated user's email address. The application validates that this email ends with the configured `permitted_email_suffix`.
