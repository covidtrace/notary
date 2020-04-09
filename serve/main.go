package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2/google"
)

func main() {
	router := httprouter.New()

	serviceAccount := os.Getenv("GOOGLE_SERVICE_ACCOUNT")
	if serviceAccount == "" {
		panic(errors.New("GOOGLE_SERVICE_ACCOUNT environment variable is required"))
	}

	bucketList := os.Getenv("CLOUD_STORAGE_BUCKETS")
	if bucketList == "" {
		panic(errors.New("STORAGE_BUCKETS environment variable is required"))
	}

	buckets := strings.Split(bucketList, ",")
	if len(buckets) == 0 {
		panic(errors.New("No buckets configured"))
	}

	router.POST("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		conf, err := google.JWTConfigFromJSON([]byte(serviceAccount))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		query := r.URL.Query()

		bucket := query.Get("bucket")
		if bucket == "" {
			bucket = buckets[0]
		} else {
			found := false
			for _, candidate := range buckets {
				if strings.EqualFold(candidate, bucket) {
					bucket = candidate
					found = true
					break
				}
			}

			if !found {
				http.Error(w, "Bucket not allowed", http.StatusBadRequest)
				return
			}
		}

		contentType := query.Get("contentType")
		if contentType == "" {
			http.Error(w, "`contentType` is a required parameter", http.StatusBadRequest)
			return
		}

		object := query.Get("object")
		if object == "" {
			http.Error(w, "`object` is a required parameter", http.StatusBadRequest)
			return
		}

		opts := &storage.SignedURLOptions{
			ContentType:    contentType,
			Expires:        time.Now().Add(15 * time.Minute),
			GoogleAccessID: conf.Email,
			Method:         "PUT",
			PrivateKey:     conf.PrivateKey,
		}

		signedURL, err := storage.SignedURL(bucket, object, opts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		m := struct {
			Status    string `json:"status"`
			SignedURL string `json:"signed_url"`
		}{
			Status:    "success",
			SignedURL: signedURL,
		}

		b, err := json.Marshal(m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		io.Copy(w, bytes.NewReader(b))
	})

	router.PanicHandler = func(w http.ResponseWriter, _ *http.Request, _ interface{}) {
		http.Error(w, "Unknown error", http.StatusBadRequest)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}
