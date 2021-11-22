package media

import (
	"fmt"

	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

type Metadata struct {
	Make         string
	Model        string
	FNumber      Fraction
	ShutterSpeed SignedFraction
	ISOSpeed     uint16
}

type Fraction struct {
	Numerator, Denominator uint32
}

type SignedFraction struct {
	Numerator, Denominator int32
}

func ExtractMetadata(b []byte) (metadata Metadata, err error) {
	rawExif, err := exif.SearchAndExtractExif(b)
	if err != nil {
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

			metadata.ShutterSpeed.Numerator = val[0].Numerator
			metadata.ShutterSpeed.Denominator = val[0].Denominator
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

		return nil
	}

	err = index.RootIfd.EnumerateTagsRecursively(cb)
	if err != nil {
		return metadata, fmt.Errorf("failed to walk exif data tree: %s", err)
	}

	return metadata, nil
}
