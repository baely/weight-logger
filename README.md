# Weight Logger

Some aggregate and automation stuff

- Apple Watch, Withings, MyFitnessPal regularly report measurements to Apple Health
- Every 45 minutes Auto Export pushes data up to `/data`
- `/data` saves pushed data to Firestore
- Cloud Function `GenerateProcessImage` listens to Firestore changes and generates respective daily images into Cloud Storage
- Cloud Scheduler hits `/post-image` at 8:30am daily which triggers a new post to Instagram with yesterday's image
- Cloud Scheduler hits `/refresh-token` at midnight on Sundays to refresh the Instagram token
