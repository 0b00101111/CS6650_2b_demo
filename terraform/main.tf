# Wire together modules: network, ecr, logging, alb, ecs, autoscaling
module "network" {
  source         = "./modules/network"
  service_name   = var.service_name
  container_port = var.container_port
}

module "ecr" {
  source          = "./modules/ecr"
  repository_name = var.ecr_repository_name
}

module "logging" {
  source            = "./modules/logging"
  service_name      = var.service_name
  retention_in_days = var.log_retention_days
}

# Reuse an existing IAM role for ECS tasks
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

# Application Load Balancer
module "alb" {
  source                 = "./modules/alb"
  service_name           = var.service_name
  container_port         = var.container_port
  vpc_id                 = module.network.vpc_id
  subnet_ids             = module.network.subnet_ids
  alb_security_group_ids = [module.network.alb_security_group_id]
}

module "ecs" {
  source             = "./modules/ecs"
  service_name       = var.service_name
  image              = "${module.ecr.repository_url}:latest"
  container_port     = var.container_port
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging.log_group_name
  ecs_count          = var.ecs_count
  region             = var.aws_region
  target_group_arn   = module.alb.target_group_arn
}

# Auto Scaling
module "autoscaling" {
  source       = "./modules/autoscaling"
  cluster_name = module.ecs.cluster_name
  service_name = module.ecs.service_name
  min_capacity = 2
  max_capacity = 4
  cpu_target   = 70
}

// Build & push the Go app image into ECR
resource "docker_image" "app" {
  name = "${module.ecr.repository_url}:latest"
  build {
    context = "../src"
  }
}

resource "docker_registry_image" "app" {
  name = docker_image.app.name
}
