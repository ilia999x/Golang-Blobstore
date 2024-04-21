package main

import (
	"context"
	"fmt"
	"image/draw"
	"mime/multipart"
	"os"
	"io"
	"time"
	"bytes"
	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"image"
	"log"
	"net/http"
    "cloud.google.com/go/storage"
	"basic_server/server/static"
	"github.com/gin-gonic/gin"
	"image/jpeg"
)

const (
	projectID  = os.Setenv("PROJECT_ID", "website-333") // FILL IN WITH YOUR FILE PATH  // FILL IN WITH YOURS
	bucketName =  os.Setenv("BUCKETNAME", "webbucket")// FILL IN WITH YOURS
	bucketNameblob = os.Setenv("BLOB_BUCKET", "blob_bucket")
)

type ClientUploader struct {
	cl         *storage.Client
	projectID  string
	bucketName string
	uploadPath string
}
type BloopUploader struct {
	cl         *storage.Client
	projectID  string
	bucketName string
	bucketNameblob string
	uploadPath string
}
var blooper *BloopUploader

var uploader *ClientUploader
func checkError(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
func init() {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "serviceAccountKey.json") // FILL IN WITH YOUR FILE PATH
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	blooper = &BloopUploader{
		cl:         client,
		bucketName: bucketNameblob,
		projectID:  projectID,
		uploadPath: "blop_images/",
	}
	uploader = &ClientUploader{
		cl:         client,
		bucketName: bucketName,
		projectID:  projectID,
		uploadPath: "post_images/",
	}
}
func main() {
  	r := gin.Default()
  	r.Use(static.Serve("/", static.LocalFile("/tmp", false)))
  	r.GET("/ping", func(c *gin.Context) {
  	  c.String(200, "test")
  	})
	r.POST("/upload", func(c *gin.Context) {
		f, err := c.FormFile("post_images")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		blobFile, err := f.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		err = uploader.UploadFile(blobFile, f.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "success",
		})
	
	})
	r.POST("/uploadblob", func(c *gin.Context) {
		f, err := c.FormFile("image_blobs")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		blobFile, err := f.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		err = blooper.blobUploadFile(blobFile, f.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(200, gin.H{
			"message": "success",
		})
	
	})
	r.GET("/post/:imageid", func(c *gin.Context) {
		imageid := c.Param("imageid")
		format := c.Query("format")
		name := c.Query("name")
		get_url := c.Query("url")
		valoriginal := uploader.valoriginalimage(imageid)
		if (valoriginal) {
			val := blooper.validatorblob(imageid,name,format)
		
			if (val==false && name!= "extra") {
				FetchAndResize, err:=FetchAndResizeImage(get_url,name)
				if err != nil {
					return
				}
				newgmg := EncodeImageToJpg(FetchAndResize,format)

				err = blooper.blobUploadFileDirect(newgmg,imageid,name,format)
				if err != nil {
				}

			}
			// Split the URL into parts using the "/" separator
			parts := strings.Split(get_url, "/")

			// Get the last part of the URL, which is the filename
			filename := parts[len(parts)-1]

			// Split the filename into name and extension
			nameParts := strings.Split(filename, ".")
			name := nameParts[0] // Name without extension
			extension := nameParts[1] // File extension

			// Modify the name (for example, add "_modified" to the name)
			modifiedName := imageid+"_"+name+"_"+format

			// Join the parts back together to form the modified URL
			modifiedURL := strings.Join(parts[:len(parts)-1], "/") + "/" + modifiedName + "." + extension
			
			response, err := http.Get(modifiedURL)
			if err != nil {
				return 
			}
			reader := response.Body
			contentLength := response.ContentLength
			contentType := response.Header.Get("Content-Type")
			extraHeaders := map[string]string{
				"Content-Disposition": `inline `,
				"file-format":format,
				"filesize":name,
				"age":" 107199",
			}

			c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
		}
		if (valoriginal==false) {
			c.JSON(404, gin.H{
				"code": "wrong image id", "message": "image in store doesnt exist",
			})
		}
		
	})

  	if err := r.Run(":7788"); err != nil {
  	  log.Fatal(err)
  	}
}
func imageToRGBA(src image.Image) *image.RGBA {

    // No conversion needed if image is an *image.RGBA.
    if dst, ok := src.(*image.RGBA); ok {
        return dst
    }

    // Use the image/draw package to convert to *image.RGBA.
    b := src.Bounds()
    dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
    draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
    return dst
}


func (c *ClientUploader ) valoriginalimage( object string) bool {
	
	ctx := context.Background()
	
	attrs,err := c.cl.Bucket(c.bucketName).Object(c.uploadPath + object+".jpg").Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		fmt.Println("The object does not exist")
		return false
	}
	if err != nil {
		// TODO: handle error.
	}
	fmt.Printf("The object exists and has attributes: %#v\n", attrs)
	return true
}


func (c *BloopUploader) validatorblob(object string, size string, format string) bool {
	
	ctx := context.Background()
	
	
	objectname := object+"_"+size+"_"+format+"."+format
	attrs,err := c.cl.Bucket(c.bucketName).Object(c.uploadPath + objectname).Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		fmt.Println("The object does not exist")
		return false
	}
	if err != nil {
		// TODO: handle error.
	}
	fmt.Printf("The object exists and has attributes: %#v\n", attrs)
	return true
}

func (c *BloopUploader) blobUploadFileDirect( img *bytes.Buffer, object string, size string, format string) error{
	
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	objectname := object+"_"+size+"_"+format+"."+format
	wc := c.cl.Bucket(c.bucketName).Object(c.uploadPath + objectname).NewWriter(ctx)
	if _, err := io.Copy(wc, img); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}
// struct containing the initial query params
type ResizerParams struct {
	url    string
}

// fetch the image from provided url and resize it
func FetchAndResizeImage(url , name string)(*image.Image, error) {
	var dst image.Image
	
	// fetch input data
	response, err := http.Get(url)
	if err != nil {
		return &dst, err
	}
	// don't forget to close the response
	defer response.Body.Close()

	// decode input data to image
	src, _, err := image.Decode(response.Body)
	if err != nil {
		return &dst, err
	}
	var main = 2
	
	if name=="medium"{
		main = 800
	}
	if name=="small"{
		main = 680
	}
	if name=="nano"{
		main = 300
	}
	if name=="micro"{
		main = 100
	}

	b := src.Bounds()
	imgWidth := b.Max.X
	imgHeight := b.Max.Y
	reimgWidth,reimgHeight  := CalcFactors(imgWidth,imgHeight,main)

	dst = imaging.Resize(src, reimgWidth, reimgHeight, imaging.Lanczos)
	// resize input image
	return &dst, nil
}

func CalcFactors(width,height, resize int)(Readywidth,Readyheight int){
	normal := false
	horisontal := true
	check := false
	
	if width >= height{
		horisontal = true
		check = true
	}

	if width < height{
		horisontal = false
	}
	if resize > height && resize > width{
		normal = true
	}

	if normal == true{
		Readywidth=width-2
		Readyheight=height-2
	}

	if horisontal == true && check == true{
		cuted := width-resize
		size := float64(height)/float64(width)
		double := size*float64(cuted)
		Readywidth= width-cuted
		Readyheight= height-int(double)
	}

	if horisontal == false{
		cuted := height-resize
		size := float64(width)/float64(height)
		double := size*float64(cuted)
		
		Readywidth=width-int(double)
		Readyheight=height-cuted
	}
	return
}

// encode image to jpeg
func EncodeImageToJpg(img *image.Image,format string) (*bytes.Buffer) {
	
	encoded := &bytes.Buffer{}
	if (format=="webp"){
		err := webp.Encode(encoded, *img, &webp.Options{Quality : 20});
		if err != nil {
			log.Println(err)
		}
	}
	if (format=="jpg"){
		err := jpeg.Encode(encoded, *img, nil)
		if err != nil {
			
			log.Println(err)
		}
	}
	return encoded
}

func (c *ClientUploader) UploadFile(file multipart.File, object string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := c.cl.Bucket(c.bucketName).Object(c.uploadPath + object).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}

func (c *BloopUploader) blobUploadFile(file multipart.File, object string)  error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	wc := c.cl.Bucket(c.bucketName).Object(c.uploadPath + object).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
	
}
