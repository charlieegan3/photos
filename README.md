# photos

This is the source code for the application running at
[photos.charlieegan3.com](https://photos.charlieegan3.com/).
The app is a Go application and runs on
[Northflank](https://northflank.com).

The project is a natural continuation of a project I built to
back up my Instagram account
[in 2018](https://charlieegan3.com/posts/2018-03-04-backing-up-instagram).

## Config

Should you want to run an instance of this application yourself, the required configuration file looks like this:

```yaml
hostname: photos.charlieegan3.com
environment: production
server:
  address: "0.0.0.0"
  port: "3000"
  adminUsername: "example"
  adminPassword: "example"
geoapify:
  url: https://maps.geoapify.com/v1/staticmap
  key: xxx
database:
  migrationsPath: file:///etc/config/migrations
  createDatabase: false
  connectionString: postgres://example:example@example.com
  params:
    dbname: example
bucket:
  # I use gocloud.dev so it should be possible to use other cloud providers
  url: gs://example-bucket
  webUrl: https://storage.googleapis.com/example-bucket/
notification_webhook:
  endpoint: https://example.com/
```