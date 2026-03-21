output "elastic_ip" {
  value       = aws_eip.pill_doser.public_ip
  description = "Public IP of the EC2 instance"
}

output "api_url" {
  value       = "http://${aws_eip.pill_doser.public_ip}:8080"
  description = "Base URL of the API"
}
