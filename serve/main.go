package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/covidtrace/utils/env"
	httputils "github.com/covidtrace/utils/http"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2/google"
)

func main() {
	router := httprouter.New()

	serviceAccount := env.MustGet("GOOGLE_SERVICE_ACCOUNT")
	bucketList := env.MustGet("CLOUD_STORAGE_BUCKETS")

	buckets := strings.Split(bucketList, ",")
	if len(buckets) == 0 {
		panic(errors.New("No buckets configured"))
	}

	router.POST("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		conf, err := google.JWTConfigFromJSON([]byte(serviceAccount))
		if err != nil {
			httputils.ReplyInternalServerError(w, err)
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
				httputils.ReplyBadRequestError(w, errors.New("Bucket not allowed"))
				return
			}
		}

		contentType := query.Get("contentType")
		if contentType == "" {
			httputils.ReplyBadRequestError(w, errors.New("`contentType` is a required parameter"))
			return
		}

		object := query.Get("object")
		if object == "" {
			httputils.ReplyBadRequestError(w, errors.New("`object` is a required parameter"))
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
			httputils.ReplyBadRequestError(w, err)
			return
		}

		httputils.ReplyJSON(w, struct {
			Status    string `json:"status"`
			SignedURL string `json:"signed_url"`
		}{
			Status:    "success",
			SignedURL: signedURL,
		}, http.StatusOK)
	})

	router.PanicHandler = func(w http.ResponseWriter, _ *http.Request, _ interface{}) {
		httputils.ReplyInternalServerError(w, errors.New("Unknown error"))
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", env.GetDefault("PORT", "8080")), router))
}
