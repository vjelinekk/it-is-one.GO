# Deployment Plan: AWS EC2 + Terraform + SES Email

## Overview

- **Infrastructure**: Terraform provisions EC2, Security Group, Elastic IP, IAM role
- **Deployment**: EC2 user_data script (runs on boot) — installs Docker, clones GitHub repo, starts container
- **Email**: AWS SES replaces `log.Printf("[EMAIL]...")` in the worker
- **Database**: SQLite persisted via Docker volume mount on EC2 disk

---

## Part 1: AWS SES Email (Go code changes)

### New dependency
```
github.com/aws/aws-sdk-go-v2/config
github.com/aws/aws-sdk-go-v2/service/ses
```

### New file: `pkg/email/ses.go`
- `Init(fromEmail string)` — loads AWS config using EC2 instance role (no hardcoded keys)
- `SendMissedDoseAlert(to, patientEmail, scheduledTime string) error` — sends email via SES

### `cmd/server/main.go`
```go
email.Init(os.Getenv("SES_FROM_EMAIL"))
```

### `pkg/server/worker.go`
Replace:
```go
log.Printf("[EMAIL] To: %s — Patient %s missed...", cg.Email, user.Email, scheduledTimeStr)
```
With:
```go
email.SendMissedDoseAlert(cg.Email, user.Email, scheduledTimeStr)
```
Falls back to `log.Printf` if `SES_FROM_EMAIL` is not set (local dev).

### `docker-compose.yaml`
```yaml
volumes:
  - ./data.db:/root/data.db
env_file:
  - .env
```

---

## Part 2: Terraform (`infra/terraform/`)

| File | Purpose |
|------|---------|
| `main.tf` | AWS provider, region (default `eu-central-1`) |
| `ec2.tf` | EC2 t3.micro, Security Group (ports 22 + 8080), Elastic IP, user_data script |
| `key_pair.tf` | Generates RSA key pair, saves `pill-doser.pem` locally, uploads public key to AWS |
| `iam.tf` | IAM role + instance profile granting `ses:SendEmail` to the EC2 instance |
| `outputs.tf` | Prints Elastic IP after apply |

### user_data script (auto-runs on EC2 first boot)
```bash
#!/bin/bash
yum update -y
yum install -y docker git
service docker start
usermod -aG docker ec2-user
curl -L https://github.com/docker/compose/releases/latest/download/docker-compose-linux-x86_64 \
  -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

git clone https://github.com/YOUR/REPO.git /app
echo "SES_FROM_EMAIL=noreply@yourdomain.com" > /app/.env
cd /app && docker-compose up -d --build
```

### `infra/Makefile`
```makefile
up:       # terraform apply — provisions EC2 + auto-deploys on boot
down:     # terraform destroy — tears everything down
redeploy: # taint EC2 + terraform apply — recreates instance, re-runs user_data
ssh:      # SSH into the EC2 instance
```

### `infra/.gitignore`
```
*.pem
*.tfstate
*.tfstate.backup
.terraform/
.terraform.lock.hcl
```

---

## SES Setup (one-time, manual)

1. AWS Console → **SES** → **Verified Identities** → verify your sender email address
2. By default SES is in **sandbox mode** — can only send to verified addresses
3. To send to anyone: request **production access** in the SES console

---

## Usage

### Prerequisites
- `aws configure` (access key + secret for your AWS account)
- [Terraform](https://developer.hashicorp.com/terraform/install) installed
- GitHub repo with this code

### Commands
```bash
cd infra

# First deploy (~3 min)
make up
# → provisions EC2, user_data installs Docker, clones repo, starts container
# → prints public IP at the end

# SSH into EC2 to check logs
make ssh

# Tear down all AWS resources
make down

# Redeploy after pushing new code to GitHub (~3 min)
make redeploy
```

---

## Files to create/modify

**Create:**
- `pkg/email/ses.go`
- `infra/terraform/main.tf`
- `infra/terraform/ec2.tf`
- `infra/terraform/key_pair.tf`
- `infra/terraform/iam.tf`
- `infra/terraform/outputs.tf`
- `infra/Makefile`
- `infra/.gitignore`

**Modify:**
- `cmd/server/main.go` — init SES
- `pkg/server/worker.go` — call `email.SendMissedDoseAlert`
- `docker-compose.yaml` — add volume + env_file
- `go.mod` / `go.sum` — AWS SDK dependency
