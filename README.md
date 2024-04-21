
# Image Blob Resizer API

This API server is designed to handle image blob resizing and storage operations. It provides endpoints for uploading image blobs, resizing them, and storing the resized versions 

in Google Cloud Storage.

## Features

- **Image Blob Upload**: Accepts image blobs and uploads them to Google Cloud Storage.
- **Image Resizing**: Resizes uploaded image blobs to predefined dimensions.
- **Google Cloud Storage Integration**: Stores resized images in Google Cloud Storage.
- **Dynamic Image URLs**: Generates dynamic URLs for accessing resized images.

## Prerequisites

Before running the server, make sure you have the following set up:

- Google Cloud Storage bucket with appropriate permissions.
- Service account key for authenticating with Google Cloud Storage.
- Environment variables set for `PROJECT_ID`, `BUCKET_NAME`, and `BLOB_BUCKET`.

## Endpoints

### Upload Image Blob

- **URL**: `/uploadblob`
- **Method**: POST
- **Request Parameters**: Form-data with the image blob to upload
- **Response**: JSON response indicating the success or failure of the upload operation

### Retrieve Resized Image Blob

- **URL**: `/post/:imageid`
- **Method**: GET
- **Request Parameters**:
  - `imageid`: The unique identifier of the resized image
  - `format`: The format of the resized image (e.g., "jpg", "webp")
  - `name`: The name of the resized image (e.g., "medium", "small")
  - `url`: The URL of the original image blob
- **Response**: The resized image blob corresponding to the provided parameters

## Usage

1. Set up your Google Cloud Storage bucket and obtain the necessary credentials.
2. Set the required environment variables (`PROJECT_ID`, `BUCKET_NAME`, `BLOB_BUCKET`, `GOOGLE_APPLICATION_CREDENTIALS`).
3. Start the server using `go run main.go`.
4. Use the provided endpoints to upload image blobs and retrieve resized images.

## Dependencies

- Gin: Web framework for building APIs in Go
- Cloud Storage: Google Cloud Storage client for interacting with storage buckets
- Imaging: Library for basic image processing tasks
- Webp: Library for encoding and decoding WebP images

