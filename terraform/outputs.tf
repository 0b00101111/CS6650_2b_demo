output "ecs_cluster_name" {
  value = module.ecs.cluster_name
}

output "ecs_service_name" {
  value = module.ecs.service_name
}

output "alb_dns_name" {
  description = "ALB DNS name - use this for load testing"
  value       = module.alb.alb_dns_name
}
