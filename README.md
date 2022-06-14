# oauth2-server

## Setup

### Environment

```sh
cp .env.example .env
```

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
