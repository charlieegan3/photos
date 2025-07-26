# photos

[photos.charlieegan3.com](https://photos.charlieegan3.com) is a personal, single user photo sharing site
which I built and host for myself. The application has a number of features I find valuable:

- [Photo Map](https://photos.charlieegan3.com/locations)
- [Trips](https://photos.charlieegan3.com/posts/period)
- [Search](https://photos.charlieegan3.com/posts/search)
- [Browse by Device](https://photos.charlieegan3.com/devices)
- [RSS](https://photos.charlieegan3.com/rss.xml)

The app is formed of a Go application. The project is the
spiritual successor of a project I built to back up my Instagram account
[in 2018](https://charlieegan3.com/posts/2018-03-04-backing-up-instagram).

## Config

Should you want to run an instance of this application yourself,
the required configuration file looks like this:

```yaml
environment: development
admin:
  auth:
    param: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    provider_url: https://xxxxxxxxxxxxxx
    client_id: xxxxxxxxxxxxxxxxxxxxxxx
    client_secret: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
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

Note that a database will also need to be provided.
