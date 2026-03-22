resource "aws_iam_role" "pill_doser" {
  name = "pill-doser-ec2-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "ses_send" {
  name = "pill-doser-ses-send"
  role = aws_iam_role.pill_doser.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["ses:SendEmail", "ses:VerifyEmailIdentity", "sns:Publish", "sns:CreateSMSSandboxPhoneNumber", "sns:VerifySMSSandboxPhoneNumber", "sms-voice:CreateVerifiedDestinationNumber", "sms-voice:SendDestinationNumberVerificationCode", "sms-voice:VerifyDestinationNumber", "sms-voice:SendTextMessage", "sms-voice:DescribeVerifiedDestinationNumbers", "logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents", "logs:DescribeLogStreams"]
      Resource = "*"
    }]
  })
}

resource "aws_iam_instance_profile" "pill_doser" {
  name = "pill-doser-instance-profile"
  role = aws_iam_role.pill_doser.name
}
