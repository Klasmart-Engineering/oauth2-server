# oauth2-server

## Setup

### Secrets

```sh
sudo chmod +x scripts/generate_secrets.sh
./scripts/generate_secrets.sh
```

### DynamoDB

```
docker-compose up
cd terraform/dev
terraform apply --auto-approve
```
