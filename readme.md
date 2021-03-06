# covidtrace/notary

Notary handles generating signed Cloud Storage PUT URLs so that the COVID
Trace app can upload files directly to Cloud Storage.

## API

### `POST /`

```
curl -XPOST 'https://covidtrace-notary.domain/?contentType=text/csv&object=sample.csv&bucket=BUCKET'
{"success": true, "signed_url": "SIGNED_CLOUD_STORAGE_PUT_URL"}
```

## Deploying

Notary is deployed as a Google Cloud Run service behind a
[rate limiting proxy](https://github.com/covidtrace/proxy). It is designed to
be easy to deploy and configure using the following environment variables.

```
GOOGLE_SERVICE_ACCOUNT="JSON service account file generated by IAM"
CLOUD_STORAGE_BUCKETS="comma separated list of buckets to sign URLs for"
```