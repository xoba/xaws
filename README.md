# xaws

`xaws` is a small collection of helpers that simplify working with the AWS SDK for Go.
It wraps a few common operations for services such as S3, SES, KMS and Step Functions.

## Quick start

1. **Install Go 1.24 or newer** and make sure your AWS credentials are available
   through the standard configuration chain (environment variables, config files, etc.).
2. **Clone the repository** and change into the directory:

   ```bash
   git clone https://github.com/xoba/xaws.git
   cd xaws
   ```
3. **Run the tests** to verify everything works:

   ```bash
   go test ./...
   ```
4. **Format the code** before committing any changes:

   ```bash
   find . -name "*.go" -exec gofmt -w {} \;
   ```
5. **Build the module** (optional) to fetch dependencies and compile:

   ```bash
   go build .
   ```

## Usage example

After importing the module you can create AWS service clients using the helper
functions provided. Here is a small example that uploads a file using S3:

```go
import (
    "context"
    "github.com/xoba/xaws"
)

func uploadFile(f xaws.File) error {
    s3Client, err := xaws.NewS3()
    if err != nil {
        return err
    }
    _, err = xaws.UploadMultipart(context.Background(), s3Client, f, "bucket", "key")
    return err
}
```

This is only a starting point&mdash;check the source for other utilities.

