TF_DIR := terraform
PEM     := pill-doser.pem

.PHONY: up down redeploy ssh deploy-rds tunnel

up:
	cd $(TF_DIR) && terraform init && terraform apply

down:
	cd $(TF_DIR) && terraform destroy

redeploy:
	cd $(TF_DIR) && terraform taint aws_instance.pill_doser && terraform apply

ssh:
	ssh -i $(PEM) -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null ec2-user@$$(cd $(TF_DIR) && terraform output -raw elastic_ip)

tunnel:
	ssh -i $(PEM) -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
		-L 5432:$$(cd $(TF_DIR) && terraform output -raw rds_endpoint) \
		ec2-user@$$(cd $(TF_DIR) && terraform output -raw elastic_ip) -N

deploy-rds:
	cd $(TF_DIR) && terraform apply
	$(eval IP := $(shell cd $(TF_DIR) && terraform output -raw elastic_ip))
	$(eval RDS := $(shell cd $(TF_DIR) && terraform output -raw rds_endpoint))
	$(eval PASS := $(shell cd $(TF_DIR) && grep db_password terraform.tfvars | cut -d'"' -f2))
	ssh -i $(PEM) -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null ec2-user@$(IP) \
		"echo 'DATABASE_URL=postgres://pilldoser:$(PASS)@$(RDS)/pilldoser' | sudo tee -a /app/.env && \
		cd /app && sudo git pull origin main && sudo docker-compose up --build -d"
