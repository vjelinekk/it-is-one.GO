data "aws_ami" "amazon_linux_2" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_security_group" "pill_doser" {
  name        = "pill-doser-sg"
  description = "Allow SSH and API traffic"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "pill_doser" {
  ami                    = data.aws_ami.amazon_linux_2.id
  instance_type          = "t3.micro"
  key_name               = aws_key_pair.pill_doser.key_name
  vpc_security_group_ids = [aws_security_group.pill_doser.id]
  iam_instance_profile   = aws_iam_instance_profile.pill_doser.name

  user_data = <<-EOF
    #!/bin/bash
    yum update -y
    yum install -y docker git
    service docker start
    usermod -aG docker ec2-user
    curl -L https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-linux-x86_64 \
      -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose

    git clone ${var.github_repo} /app
    echo "SES_FROM_EMAIL=${var.ses_from_email}" > /app/.env
    touch /app/data.db
    cd /app && docker-compose up -d --build
  EOF

  tags = {
    Name = "pill-doser"
  }
}

resource "aws_eip" "pill_doser" {
  instance = aws_instance.pill_doser.id
  domain   = "vpc"
}
