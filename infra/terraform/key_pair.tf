resource "tls_private_key" "pill_doser" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "pill_doser" {
  key_name   = "pill-doser"
  public_key = tls_private_key.pill_doser.public_key_openssh
}

resource "local_file" "private_key" {
  content         = tls_private_key.pill_doser.private_key_pem
  filename        = "${path.module}/../pill-doser.pem"
  file_permission = "0600"
}
