# photos

[photos.charlieegan3.com](https://photos.charlieegan3.com) is a personal, single user photo sharing site
which I built and host for myself. The application has a number of features I find valuable:

* [Photo Map](https://photos.charlieegan3.com/locations)
* [Trips](https://photos.charlieegan3.com/posts/period)
* [Search](https://photos.charlieegan3.com/posts/search)
* [Browse by Device](https://photos.charlieegan3.com/devices)
* [RSS](https://photos.charlieegan3.com/rss.xml)

The app is formed of a Go application and runs on
[Northflank](https://northflank.com). The project is the
spiritual successor of a project I built to back up my Instagram account
[in 2018](https://charlieegan3.com/posts/2018-03-04-backing-up-instagram).

## Config

Should you want to run an instance of this application yourself,
the required configuration file looks like this:

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
bucket:
  # I use gocloud.dev, so it should be possible to use various cloud providers
  url: gs://example-bucket # file:///tmp/photos
notification_webhook:
  endpoint: https://example.com/
```

Note that a database will also need to be provided.
