package mediametadata

import (
	"fmt"
	"time"

	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

type Metadata struct {
	Make  string
	Model string

	DateTime time.Time

	FNumber      Fraction
	ShutterSpeed Fraction
	ISOSpeed     uint16

	Latitude  Coordinate
	Longitude Coordinate
	Altitude  Altitude
}

type Coordinate struct {
	Degrees Fraction
	Minutes Fraction
	Seconds Fraction

	Ref string
}

func (c *Coordinate) ToDecimal() (float64, error) {
	multiplier := 1.0
	if c.Ref == "S" || c.Ref == "W" {
		multiplier = -1.0
	}

	degrees, err := c.Degrees.ToDecimal()
	if err != nil {
		return 0, fmt.Errorf("coordinate can't be converted to decimal: %s", err)
	}
	minutes, err := c.Minutes.ToDecimal()
	if err != nil {
		return 0, fmt.Errorf("coordinate can't be converted to decimal: %s", err)
	}
	seconds, err := c.Seconds.ToDecimal()
	if err != nil {
		return 0, fmt.Errorf("coordinate can't be converted to decimal: %s", err)
	}

	return (degrees + minutes/float64(60) + seconds/float64(3600)) * multiplier, nil
}

type Altitude struct {
	Value Fraction
	Ref   byte
}

func (a *Altitude) ToDecimal() (float64, error) {
	value, err := a.Value.ToDecimal()
	if err != nil {
		return 0, fmt.Errorf("altitude can't be converted to decimal: %s", err)
	}

	multiplier := 1.0
	if a.Ref == 1 {
		multiplier = -1.0
	}

	return value * multiplier, nil
}

type Fraction struct {
	Numerator, Denominator uint32
}

func (f *Fraction) ToDecimal() (float64, error) {
	if f.Denominator == 0 {
		return 0, fmt.Errorf("fraction with 0 denominator cannot be converted to decimal")
	}

	return float64(f.Numerator) / float64(f.Denominator), nil
}

func ExtractMetadata(b []byte) (metadata Metadata, err error) {
	rawExif, err := exif.SearchAndExtractExif(b)
	if err == exif.ErrNoExif {
		return metadata, nil
	} else if err != nil {
		return metadata, fmt.Errorf("failed to get raw exif data: %s", err)
	}

	im, err := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		return metadata, fmt.Errorf("failed to create idfmapping: %s", err)
	}

	ti := exif.NewTagIndex()

	_, index, err := exif.Collect(im, ti, rawExif)
	if err != nil {
		return metadata, fmt.Errorf("failed to collect exif data: %s", err)
	}

	cb := func(ifd *exif.Ifd, ite *exif.IfdTagEntry) error {
		if ite.TagName() == "Make" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw Make value")
			}

			val, ok := rawValue.(string)
			if !ok {
				return fmt.Errorf("Make was not in expected format: %#v", rawValue)
			}

			metadata.Make = val
		}

		if ite.TagName() == "Model" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw Model value")
			}

			val, ok := rawValue.(string)
			if !ok {
				return fmt.Errorf("Model was not in expected format: %#v", rawValue)
			}

			metadata.Model = val
		}

		if ite.TagName() == "DateTimeOriginal" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw DateTimeOriginal value")
			}

			val, ok := rawValue.(string)
			if !ok {
				return fmt.Errorf("DateTimeOriginal was not in expected format: %#v", rawValue)
			}

			metadata.DateTime, err = time.Parse("2006:01:02 15:04:05", val)
			if err != nil {
				return fmt.Errorf("failed to parse time: %s", err)
			}
		}

		if ite.TagName() == "FNumber" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw FNumber value")
			}
			val, ok := rawValue.([]exifcommon.Rational)
			if !ok {
				return fmt.Errorf("FNumber was not in expected format: %#v", rawValue)
			}

			if len(val) != 1 {
				return fmt.Errorf("found %d FNumbers", len(val))
			}

			metadata.FNumber.Numerator = val[0].Numerator
			metadata.FNumber.Denominator = val[0].Denominator
		}

		if ite.TagName() == "ShutterSpeedValue" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw ShutterSpeed value")
			}

			val, ok := rawValue.([]exifcommon.SignedRational)
			if !ok {
				return fmt.Errorf("ShutterSpeed was not in expected format: %#v", rawValue)
			}

			if len(val) != 1 {
				return fmt.Errorf("found %d ShutterSpeedValues", len(val))
			}

			metadata.ShutterSpeed.Numerator = uint32(val[0].Numerator)
			metadata.ShutterSpeed.Denominator = uint32(val[0].Denominator)
		}

		if ite.TagName() == "ISOSpeedRatings" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw ISOSpeedRatings value")
			}

			val, ok := rawValue.([]uint16)
			if !ok {
				return fmt.Errorf("ISOSpeedRatings was not in expected format: %#v", rawValue)
			}

			if len(val) != 1 {
				return fmt.Errorf("found %d ISOSpeedRatings", len(val))
			}

			metadata.ISOSpeed = val[0]
		}

		if ite.TagName() == "GPSLatitudeRef" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw GPSLatitudeRef value")
			}

			val, ok := rawValue.(string)
			if !ok {
				return fmt.Errorf("GPSLatitudeRef was not in expected format: %#v", rawValue)
			}

			metadata.Latitude.Ref = val
		}

		if ite.TagName() == "GPSLatitude" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw GPSLatitude value")
			}

			val, ok := rawValue.([]exifcommon.Rational)
			if !ok {
				return fmt.Errorf("GPSLatitude was not in expected format: %#v", rawValue)
			}

			if len(val) != 3 {
				return fmt.Errorf("found %d GPSLatitude", len(val))
			}

			metadata.Latitude.Degrees.Numerator = val[0].Numerator
			metadata.Latitude.Degrees.Denominator = val[0].Denominator
			metadata.Latitude.Minutes.Numerator = val[1].Numerator
			metadata.Latitude.Minutes.Denominator = val[1].Denominator
			metadata.Latitude.Seconds.Numerator = val[2].Numerator
			metadata.Latitude.Seconds.Denominator = val[2].Denominator
		}

		if ite.TagName() == "GPSLongitudeRef" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw GPSLongitudeRef value")
			}

			val, ok := rawValue.(string)
			if !ok {
				return fmt.Errorf("GPSLongitudeRef was not in expected format: %#v", rawValue)
			}

			metadata.Longitude.Ref = val
		}

		if ite.TagName() == "GPSLongitude" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw GPSLongitude value")
			}

			val, ok := rawValue.([]exifcommon.Rational)
			if !ok {
				return fmt.Errorf("GPSLongitude was not in expected format: %#v", rawValue)
			}

			if len(val) != 3 {
				return fmt.Errorf("found %d GPSLongitude", len(val))
			}

			metadata.Longitude.Degrees.Numerator = val[0].Numerator
			metadata.Longitude.Degrees.Denominator = val[0].Denominator
			metadata.Longitude.Minutes.Numerator = val[1].Numerator
			metadata.Longitude.Minutes.Denominator = val[1].Denominator
			metadata.Longitude.Seconds.Numerator = val[2].Numerator
			metadata.Longitude.Seconds.Denominator = val[2].Denominator
		}

		if ite.TagName() == "GPSAltitudeRef" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw GPSAltitudeRef value")
			}

			val, ok := rawValue.([]byte)
			if !ok {
				return fmt.Errorf("GPSAltitudeRef was not in expected format: %#v", rawValue)
			}

			if len(val) != 1 {
				return fmt.Errorf("found %d GPSAltitudeRef", len(val))
			}

			metadata.Altitude.Ref = val[0]
		}

		if ite.TagName() == "GPSAltitude" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw GPSAltitude value")
			}

			val, ok := rawValue.([]exifcommon.Rational)
			if !ok {
				return fmt.Errorf("GPSAltitude was not in expected format: %#v", rawValue)
			}

			if len(val) != 1 {
				return fmt.Errorf("found %d GPSAltitude", len(val))
			}

			metadata.Altitude.Value.Numerator = val[0].Numerator
			metadata.Altitude.Value.Denominator = val[0].Denominator
		}

		return nil
	}

	err = index.RootIfd.EnumerateTagsRecursively(cb)
	if err != nil {
		return metadata, fmt.Errorf("failed to walk exif data tree: %s", err)
	}

	return metadata, nil
}