package imageproxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"gocloud.dev/blob"
	"willnorris.com/go/imageproxy"
)

type Resizer struct {
	mu sync.Mutex
}

// ResizeInBucket resizes an image in a bucket and saves it to a new path. One at a time to reduce load on the server.
// When a batch of images are uploaded, there are many that need to be resized. This function is a throttler to ensure
// that only one image is resized at a time and the RAM usage is contained. Original images are quite large and quickly
// consume all available RAM leading to OOMKills.
func (ir *Resizer) ResizeInBucket(
	ctx context.Context,
	bucket *blob.Bucket,
	originalMediaPath string,
	imageResizeString string,
	thumbMediaPath string,
) error {
	ir.mu.Lock()
	defer ir.mu.Unlock()

	// create a reader to get the full size media from the bucket
	br, err := bucket.NewReader(ctx, originalMediaPath, nil)
	if err != nil {
		return fmt.Errorf("failed to create reader for original media: %w", err)
	}
	defer br.Close()

	// read the full size item
	buf := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buf, br)
	if err != nil {
		return fmt.Errorf("failed to copy original media into buffer: %w", err)
	}

	err = br.Close()
	if err != nil {
		return fmt.Errorf("failed to close bucket reader: %w", err)
	}

	// resize the image based on the current settings
	imageOptions := imageproxy.ParseOptions(imageResizeString)
	imageOptions.ScaleUp = false // don't attempt to make images larger if not possible

	imageBytes, err := imageproxy.Transform(buf.Bytes(), imageOptions)
	buf = bytes.NewBuffer(imageBytes)

	// create a writer for the new thumb
	bw, err := bucket.NewWriter(ctx, thumbMediaPath, nil)
	if err != nil {
		return fmt.Errorf("failed to create writer for thumb media: %w", err)
	}
	defer bw.Close()

	_, err = io.Copy(bw, bytes.NewReader(imageBytes))
	if err != nil {
		return fmt.Errorf("failed to copy thumb media into bucket: %w", err)
	}
	err = bw.Close()
	if err != nil {
		return fmt.Errorf("failed to close bucket writer: %w", err)
	}

	return nil
}

func (ir *Resizer) CreateThumbInBucket(
	ctx context.Context,
	reader io.Reader,
	bucket *blob.Bucket,
	imageResizeString string,
	thumbMediaPath string,
) ([]byte, error) {
	var err error
	ir.mu.Lock()
	defer ir.mu.Unlock()

	// read the full size item
	buf := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buf, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy original media into buffer: %w", err)
	}

	// resize the image based on the current settings
	imageOptions := imageproxy.ParseOptions(imageResizeString)
	imageOptions.ScaleUp = false // don't attempt to make images larger if not possible

	imageBytes, err := imageproxy.Transform(buf.Bytes(), imageOptions)
	buf = bytes.NewBuffer(imageBytes)

	// create a writer for the new thumb
	bw, err := bucket.NewWriter(ctx, thumbMediaPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create writer for thumb media: %w", err)
	}
	defer bw.Close()

	_, err = io.Copy(bw, bytes.NewReader(imageBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to copy thumb media into bucket: %w", err)
	}

	err = bw.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close bucket writer: %w", err)
	}

	return imageBytes, nil
}
