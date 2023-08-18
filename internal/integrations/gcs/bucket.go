package gcs

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
)

func UploadFile(bucket string, object string, file io.Reader) error {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	bkt := client.Bucket(bucket)
	obj := bkt.Object(object)
	wc := obj.NewWriter(ctx)

	if _, err = io.Copy(wc, file); err != nil {
		return err
	}

	if err = wc.Close(); err != nil {
		return err
	}

	return nil
}

func ReadFile(bucket string, object string) (io.Reader, error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	bkt := client.Bucket(bucket)
	obj := bkt.Object(object)
	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}

	return r, nil
}
