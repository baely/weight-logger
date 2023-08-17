docker build -t server --platform linux/amd64  .
docker tag server gcr.io/baileybutler-syd/weightloss/server
docker push gcr.io/baileybutler-syd/weightloss/server
