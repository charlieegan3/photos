module github.com/charlieegan3/photos/cms

// +heroku goVersion go1.17
go 1.17

require (
	github.com/doug-martin/goqu/v9 v9.16.0
	github.com/dsoprea/go-exif/v3 v3.0.0-20210625224831-a6301f85c82b
	github.com/gobuffalo/plush v3.8.3+incompatible
	github.com/golang-migrate/migrate/v4 v4.15.1
	github.com/gomarkdown/markdown v0.0.0-20211212230626-5af6ad2f47df
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.2.0
	github.com/lib/pq v1.10.6
	github.com/maxatome/go-testdeep v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.8.0
	gocloud.dev v0.24.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
	willnorris.com/go/imageproxy v0.11.2
)

require (
	cloud.google.com/go v0.94.0 // indirect
	cloud.google.com/go/storage v1.16.1 // indirect
	github.com/BurntSushi/toml v1.1.0 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/client9/misspell v0.3.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/dsoprea/go-logging v0.0.0-20200710184922-b02d349568dd // indirect
	github.com/dsoprea/go-utility/v2 v2.0.0-20200717064901-2fccff4aa15e // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/fcjr/aia-transport-go v1.2.2 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-errors/errors v1.4.1 // indirect
	github.com/gobuffalo/flect v0.2.4 // indirect
	github.com/gobuffalo/github_flavored_markdown v1.1.1 // indirect
	github.com/gobuffalo/helpers v0.6.4 // indirect
	github.com/gobuffalo/tags v2.1.7+incompatible // indirect
	github.com/gobuffalo/tags/v3 v3.1.2 // indirect
	github.com/gobuffalo/validate/v3 v3.3.1 // indirect
	github.com/gofrs/uuid v4.2.0+incompatible // indirect
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/wire v0.5.0 // indirect
	github.com/googleapis/gax-go/v2 v2.1.0 // indirect
	github.com/gordonklaus/ineffassign v0.0.0-20210914165742-4cc7213b9bc8 // indirect
	github.com/gorilla/css v1.0.0 // indirect
	github.com/gorilla/feeds v1.1.1 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/kisielk/errcheck v1.6.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mdempsky/unconvert v0.0.0-20200228143138-95ecdbfc0b5f // indirect
	github.com/microcosm-cc/bluemonday v1.0.17 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/muesli/smartcrop v0.3.0 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646 // indirect
	github.com/pelletier/go-toml v1.9.3 // indirect
	github.com/philhofer/tcx v0.0.0-20140422172628-84b8274e31dc // indirect
	github.com/philhofer/vec v0.0.0-20140421144027-536fc796d369 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/rwcarlsen/goexif v0.0.0-20190401172101-9e8deecbddbd // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/sourcegraph/annotate v0.0.0-20160123013949-f4cad6c6324d // indirect
	github.com/sourcegraph/syntaxhighlight v0.0.0-20170531221838-bd320f5d308e // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tkrajina/gpxgo v1.2.1 // indirect
	github.com/tormoder/fit v0.13.0 // indirect
	github.com/twpayne/go-geom v1.4.2 // indirect
	github.com/twpayne/go-gpx v1.2.1 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/exp/typeparams v0.0.0-20220428152302-39d4317da171 // indirect
	golang.org/x/image v0.0.0-20210216034530-4410531fe030 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3 // indirect
	golang.org/x/net v0.0.0-20220802222814-0bcc04d9c69b // indirect
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f // indirect
	golang.org/x/sys v0.0.0-20220728004956-3c1f35247d10 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.11-0.20220316014157-77aa08bb151a // indirect
	golang.org/x/xerrors v0.0.0-20220411194840-2f41105eb62f // indirect
	google.golang.org/api v0.56.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20211013025323-ce878158c4d4 // indirect
	google.golang.org/grpc v1.41.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	honnef.co/go/tools v0.3.1 // indirect
	mvdan.cc/gofumpt v0.0.0-20190427031043-b295af7664cd // indirect
	willnorris.com/go/gifresize v1.0.0 // indirect
)
