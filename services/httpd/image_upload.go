package httpd

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb/models"

	"github.com/mark-rushakoff/influx-blob/blob"
	"github.com/mark-rushakoff/influx-blob/engine"
)

const imagePrefix = "whatever"

// debugUpFile helps debugging of an uploaded file.
func debugUpFile(w http.ResponseWriter, r io.Reader, header *multipart.FileHeader) {
	dst, err := os.Create("/tmp/foo.gif" + header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	//copy the uploaded file to the destination file
	if _, err := io.Copy(dst, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) serveImage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.serveImageDownload(w, r)
		return
	}
	h.serveImageUpload(w, r)
}

func (h *Handler) serveImageUpload(w http.ResponseWriter, r *http.Request) {
	fd, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer fd.Close()

	fm, err := blob.NewFileMeta(fd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fm.Path = fmt.Sprintf("%s-%d", imagePrefix, time.Now().UnixNano()/int64(time.Millisecond))
	fm.BlockSize = 1024
	fm.Time = time.Now().Unix()

	// TODO(edd): don't allocate these on each request
	v := blob.NewInfluxVolume("http://localhost:8086", "images", "")
	e := engine.NewEngine(0, 0)

	ctx := e.UploadFile(fd, fm, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	ctx.Wait()
	stats := ctx.Stats()
	uploaders, _ := e.NumWorkers()
	fmt.Printf("Uploaded %d bytes in %.2fs\n", stats.Bytes, stats.Duration.Seconds())
	fmt.Printf("(Used %d uploaders and %d chunks of %dB each)\n", uploaders, fm.NumBlocks(), fm.BlockSize)
}

// serveImageDownload64 requests files between timestamps from influx, and returns base64
// encoded versions.
func (h *Handler) serveImageDownload(w http.ResponseWriter, r *http.Request) {
	from, to := r.URL.Query().Get("from"), r.URL.Query().Get("to")
	if from == "" {
		from = fmt.Sprint(models.MinNanoTime / int64(time.Millisecond))
	}
	if to == "" {
		to = fmt.Sprint(models.MaxNanoTime / int64(time.Millisecond))
	}

	// Use list files to get all file names for a certain time range.
	v := blob.NewInfluxVolume("http://localhost:8086", "images", "")
	files, err := listFiles(from, to, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprintln(w, files)

	// Step 4. If it is then include base64 in output.
}

// func base64File(name string, v *blob.InfluxVolume) (string, error) {
// 	bms, err := v.ListBlocks(name)
// 	if err != nil {
// 		return "", err
// 	}
// 	if len(bms) == 0 {
// 		return "", fmt.Errorf("No file found for path: %s", name)
// 	}

// 	var buf bytes.Buffer
// 	e := engine.NewEngine(0, 0)
// 	ctx, err := e.DownloadFile(&buf, bms, v)
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Println("Get initiated, waiting for completion.")
// 	ctx.Wait()
// 	fmt.Println("Get complete!")

// 	fm := bms[0].FileMeta
// 	fmt.Println("Comparing checksum...")
// 	if err := fm.CompareSHA256Against(out); err != nil {
// 		return err
// 	}
// 	fmt.Println("Checksum matches. Get successful.")

// 	stats := ctx.Stats()
// 	_, downloaders := e.NumWorkers()
// 	fmt.Printf("Downloaded %d bytes in %.2fs\n", stats.Bytes, stats.Duration.Seconds())
// 	fmt.Printf("(Used %d downloaders and %d chunks of %dB each)\n", downloaders, fm.NumBlocks(), fm.BlockSize)

// 	return nil
// }

func listFiles(fromStr, toStr string, v *blob.InfluxVolume) ([]string, error) {
	var out []string

	files, err := v.ListFiles("whatever", blob.ListOptions{
		ListMatch: blob.ByPrefix,
	})
	if err != nil {
		return nil, err
	}

	from, err := strconv.ParseInt(fromStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse %q", fromStr)
	}

	to, err := strconv.ParseInt(toStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse %q", toStr)
	}

	for _, f := range files {
		split := strings.Split(f, "-")
		if len(split) < 2 {
			continue
		}

		ts, err := strconv.ParseInt(split[1], 10, 64)
		if err != nil {
			fmt.Printf("could not parse %q\n", split[1])
			continue
		}

		if ts >= from && ts <= to {
			out = append(out, f)
		} else {
			fmt.Println(out, from, to)
		}
	}

	return out, nil
}
