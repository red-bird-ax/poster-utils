package storage

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "os"

    "cloud.google.com/go/storage"
)

const ApiURL = "https://storage.cloud.google.com"

type Storage struct {
    client     *storage.Client
    bucketName string
}

func New(bucketName, projectID string) (*Storage, error) {
    ctx := context.Background()

    client, err := storage.NewClient(ctx)
    if err != nil {
        return nil, err
    }

    bucket := client.Bucket(bucketName)

    switch _, err := bucket.Attrs(ctx); err {
    case storage.ErrBucketNotExist:
        _, _ = fmt.Fprintf(os.Stderr,"[Cloud Storage] Creating bucket '%s'...\n", bucketName)
        if err = bucket.Create(ctx, projectID, nil); err != nil {
            return nil, err
        }
        fallthrough

    case nil:
        return &Storage{
            client:     client,
            bucketName: bucketName,
        }, nil

    default:
        _, _ = fmt.Fprintf(os.Stderr,"[Cloud Storage] Error %T (%s)\n", err, err.Error())
        return nil, err
    }
}

func (fs *Storage) Get(url string) ([]byte, error) {
    object := fs.client.
        Bucket(fs.bucketName).
        Object(fs.parseURL(url)).
        If(storage.Conditions{DoesNotExist: false})

    reader, err := object.NewReader(context.Background())
    if err != nil {
        return nil, err
    }

    defer func(reader *storage.Reader) {
        _ = reader.Close()
    }(reader)

    return io.ReadAll(reader)
}

func (fs *Storage) Save(content []byte, path string) (string, error) {
    return fs.upload(content, path, false)
}

func (fs *Storage) Update(content []byte, path string) error {
    _, err := fs.upload(content, path, true)
    return err
}

func (fs *Storage) Delete(url string) error {
    object := fs.client.
        Bucket(fs.bucketName).
        Object(fs.parseURL(url)).
        If(storage.Conditions{DoesNotExist: false})

    return object.Delete(context.Background())
}

func (fs *Storage) pathToURL(path string) string {
    return fmt.Sprintf("%s/%s/%s", ApiURL, fs.bucketName, path)
}

func (fs *Storage) parseURL(url string) string {
    bucketPath := fmt.Sprintf("%s/%s/", ApiURL, fs.bucketName)
    return url[len(bucketPath):]
}

func (fs *Storage) upload(content []byte, path string, update bool) (string, error) {
    ctx := context.Background()

    object := fs.client.Bucket(fs.bucketName).Object(path).If(storage.Conditions{DoesNotExist: !update})
    writer := object.NewWriter(ctx)
    buffer := bytes.NewBuffer(content)

    defer func(writer *storage.Writer) {
        _ = writer.Close()
    }(writer)

    if num, err := io.Copy(writer, buffer); err != nil {
        return "", err
    } else {
        fmt.Printf("[Cloud Storage] Updoaded %s %d bytes\n", path, num)
        if err = writer.Close(); err != nil {
            return "", err
        }
        return fs.pathToURL(object.ObjectName()), nil
    }
}