package mediametadata

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"strconv"
	"strings"
	"time"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

type Metadata struct {
	Make  string
	Model string

	Lens        string
	FocalLength string

	DateTime time.Time

	FNumber      Fraction
	ExposureTime Fraction
	ISOSpeed     uint16

	Latitude  Coordinate
	Longitude Coordinate
	Altitude  Altitude

	Height int
	Width  int
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
		return 0, fmt.Errorf("coordinate can't be converted to decimal: %w", err)
	}
	minutes, err := c.Minutes.ToDecimal()
	if err != nil {
		return 0, fmt.Errorf("coordinate can't be converted to decimal: %w", err)
	}
	seconds, err := c.Seconds.ToDecimal()
	if err != nil {
		return 0, fmt.Errorf("coordinate can't be converted to decimal: %w", err)
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
		return 0, fmt.Errorf("altitude can't be converted to decimal: %w", err)
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
		return 0, errors.New("fraction with 0 denominator cannot be converted to decimal")
	}

	return float64(f.Numerator) / float64(f.Denominator), nil
}

//nolint:maintidx
func ExtractMetadata(b []byte) (metadata Metadata, err error) {
	rawExif, err := exif.SearchAndExtractExif(b)
	if errors.Is(err, exif.ErrNoExif) {
		return metadata, nil
	} else if err != nil {
		return metadata, fmt.Errorf("failed to get raw exif data: %w", err)
	}

	im, err := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		return metadata, fmt.Errorf("failed to create idfmapping: %w", err)
	}

	ti := exif.NewTagIndex()

	_, index, err := exif.Collect(im, ti, rawExif)
	if err != nil {
		return metadata, fmt.Errorf("failed to collect exif data: %w", err)
	}

	var focalLength string
	var focalLength35mm string

	var offsetTimeOriginal string
	cb := func(_ *exif.Ifd, ite *exif.IfdTagEntry) error {
		if ite.TagName() == "Make" {
			rawValue, err := ite.Value()
			if err != nil {
				return errors.New("could not get raw Make value")
			}

			val, ok := rawValue.(string)
			if !ok {
				return fmt.Errorf("make was not in expected format: %#v", rawValue)
			}

			metadata.Make = val
		}

		if ite.TagName() == "Model" {
			rawValue, err := ite.Value()
			if err != nil {
				return errors.New("could not get raw Model value")
			}

			val, ok := rawValue.(string)
			if !ok {
				return fmt.Errorf("model was not in expected format: %#v", rawValue)
			}

			metadata.Model = val
		}

		if ite.TagName() == "LensModel" {
			rawValue, err := ite.Value()
			if err != nil {
				return errors.New("could not get raw Make value")
			}

			val, ok := rawValue.(string)
			if !ok {
				return fmt.Errorf("lens was not in expected format: %#v", rawValue)
			}

			metadata.Lens = val
		}

		if ite.TagName() == "FocalLengthIn35mmFilm" {
			rawValue, err := ite.Value()
			if err != nil {
				return errors.New("could not get raw FocalLengthIn35mmFilm value")
			}

			val, ok := rawValue.([]uint16)
			if !ok {
				return fmt.Errorf("FocalLengthIn35mmFilm was not in expected format: %#v", rawValue)
			}

			if len(val) == 1 {
				focalLength35mm = strconv.FormatUint(uint64(val[0]), 10)
			}
		}

		if ite.TagName() == "FocalLength" {
			rawValue, err := ite.Value()
			if err != nil {
				return errors.New("could not get raw FocalLength value")
			}

			val, ok := rawValue.([]exifcommon.Rational)
			if !ok {
				return fmt.Errorf("FocalLength was not in expected format: %#v", rawValue)
			}

			if len(val) == 1 {
				value := float64(val[0].Numerator) / float64(val[0].Denominator)

				focalLength = fmt.Sprintf("%.2f", value)
				focalLength = strings.TrimSuffix(focalLength, ".00")
				focalLength = strings.TrimSuffix(focalLength, "0")
			}
		}

		if ite.TagName() == "DateTimeOriginal" {
			rawValue, err := ite.Value()
			if err != nil {
				return errors.New("could not get raw DateTimeOriginal value")
			}

			val, ok := rawValue.(string)
			if !ok {
				return fmt.Errorf("DateTimeOriginal was not in expected format: %#v", rawValue)
			}

			metadata.DateTime, err = time.Parse("2006:01:02 15:04:05", val)
			if err != nil {
				return fmt.Errorf("failed to parse time: %w", err)
			}
		}

		if ite.TagName() == "OffsetTimeOriginal" {
			rawValue, err := ite.Value()
			if err != nil {
				return errors.New("could not get raw OffsetTimeOriginal value")
			}

			var ok bool
			offsetTimeOriginal, ok = rawValue.(string)
			if !ok {
				return fmt.Errorf("OffsetTimeOriginal was not in expected format: %#v", rawValue)
			}
		}

		if ite.TagName() == "FNumber" {
			rawValue, err := ite.Value()
			if err != nil {
				return errors.New("could not get raw FNumber value")
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

		if ite.TagName() == "ExposureTime" {
			rawValue, err := ite.Value()
			if err != nil {
				return errors.New("could not get raw ExposureTime value")
			}

			val, ok := rawValue.([]exifcommon.Rational)
			if !ok {
				return fmt.Errorf("ExposureTime was not in expected format: %#v", rawValue)
			}

			if len(val) != 1 {
				return fmt.Errorf("found %d ExposureTime values", len(val))
			}

			metadata.ExposureTime.Numerator = val[0].Numerator
			metadata.ExposureTime.Denominator = val[0].Denominator
		}

		if ite.TagName() == "ISOSpeedRatings" {
			rawValue, err := ite.Value()
			if err != nil {
				return errors.New("could not get raw ISOSpeedRatings value")
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
				return errors.New("could not get raw GPSLatitudeRef value")
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
				return errors.New("could not get raw GPSLatitude value")
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
				return errors.New("could not get raw GPSLongitudeRef value")
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
				return errors.New("could not get raw GPSLongitude value")
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
				return errors.New("could not get raw GPSAltitudeRef value")
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
				return errors.New("could not get raw GPSAltitude value")
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
		return metadata, fmt.Errorf("failed to walk exif data tree: %w", err)
	}

	if focalLength != "" {
		metadata.FocalLength = focalLength + "mm"

		if focalLength35mm != "" {
			metadata.FocalLength = fmt.Sprintf("%smm (%smm in 35mm format)", focalLength, focalLength35mm)
		}
	} else if focalLength35mm != "" {
		if focalLength35mm != "" {
			metadata.FocalLength = focalLength35mm + "mm in 35mm format"
		}
	}

	// handle non UTC images
	if offsetTimeOriginal != "" {
		offsetSign := 1
		if strings.HasPrefix(offsetTimeOriginal, "+") {
			// this is the reverse operation, so when the offset is +, we must subtract time to get back to UTC
			offsetSign = -1
		}
		o := strings.TrimPrefix(strings.TrimPrefix(offsetTimeOriginal, "+"), "-")
		parts := strings.Split(o, ":")
		if len(parts) == 2 {
			durationString := fmt.Sprintf("%sh%sm", parts[0], parts[1])
			duration, err := time.ParseDuration(durationString)
			if err == nil {
				metadata.DateTime = metadata.DateTime.Add(duration * time.Duration(offsetSign))
			}
		}
	}

	// special case for X100F which does not set 35mm equiv focal length
	if metadata.Make == "FUJIFILM" && metadata.Model == "X100F" {
		if metadata.FocalLength == "" || metadata.FocalLength == "23mm" {
			metadata.FocalLength = "23mm (35mm in 35mm format)"
		}
		if metadata.Lens == "" {
			metadata.Lens = "FUJINON single focal length lens"
		}
	}

	m, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return metadata, fmt.Errorf("failed to decode image for size check: %w", err)
	}
	bounds := m.Bounds()
	metadata.Width = bounds.Dx()
	metadata.Height = bounds.Dy()

	return metadata, nil
}
