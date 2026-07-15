output "instance_public_ip" {
  description = "Public IP of the EC2 instance"
  value = length(data.aws_instances.existing.ids) > 0 ? data.aws_instances.existing.public_ips[0] : aws_instance.app[0].public_ip
}

output "instance_id" {
  description = "ID of the EC2 instance"
  value = length(data.aws_instances.existing.ids) > 0 ? data.aws_instances.existing.ids[0] : aws_instance.app[0].id 
}