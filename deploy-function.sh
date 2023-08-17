gcloud functions deploy weight-gen-image \
  --gen2 \
  --runtime=go120 \
  --region=australia-southeast1 \
  --trigger-location=australia-southeast1 \
  --source=. \
  --entry-point=GenerateProgressImage \
  --trigger-event-filters=type=google.cloud.firestore.document.v1.written \
  --trigger-event-filters=database='(default)' \
  --trigger-event-filters-path-pattern=document='weightlog/{date}'
