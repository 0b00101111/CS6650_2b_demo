# CS6650 Assignment 2b - Product API with Terraform ECS Deployment

## Overview

A Product API implementation based on the provided OpenAPI specification, deployed to AWS ECS Fargate using Terraform infrastructure as code.

## Project Structure

```
.
├── src/
│   ├── main.go          # Product API server (Go + Gin)
│   ├── Dockerfile        # Multi-stage Docker build
│   ├── go.mod
│   └── go.sum
├── terraform/
│   ├── main.tf           # Module wiring + Docker build/push
│   ├── variables.tf      # Configurable variables
│   ├── provider.tf       # AWS + Docker providers
│   ├── outputs.tf        # Cluster/service names
│   └── modules/
│       ├── ecr/          # Container registry
│       ├── ecs/          # ECS cluster, task, service (Fargate)
│       ├── logging/      # CloudWatch log group
│       └── network/      # VPC, subnets, security group
├── locustfile.py         # Locust stress test
└── README.MD
```

## API Endpoints

Implements the **Product** portion of the OpenAPI spec (`api.yaml`):

| Method | Endpoint | Description | Success Code |
|--------|----------|-------------|--------------|
| GET | `/products/{productId}` | Get product by ID | 200 |
| POST | `/products/{productId}/details` | Add/update product details | 204 |

### Product Schema

```json
{
  "product_id": 1,
  "sku": "ABC-123-XYZ",
  "manufacturer": "Acme Corporation",
  "category_id": 456,
  "weight": 1250,
  "some_other_id": 789
}
```

## Deployment Instructions

### Prerequisites

- AWS CLI configured with valid credentials
- Terraform installed (v1.0+)
- Docker running locally

### Steps

1. **Clone the repository:**
   ```bash
   git clone https://github.com/0b00101111/CS6650_2b_demo.git
   cd CS6650_2b_demo
   ```

2. **Configure AWS credentials:**
   ```bash
   aws configure
   # Or for Learner Lab, set credentials in ~/.aws/credentials
   ```

3. **Deploy with Terraform:**
   ```bash
   cd terraform
   terraform init
   terraform apply
   ```
   Type `yes` when prompted. This will:
   - Create ECR repository
   - Build and push Docker image
   - Create ECS cluster, task definition, and service on Fargate
   - Set up CloudWatch logging and networking

4. **Find the public IP:**
   ```bash
   # Get task ARN
   aws ecs list-tasks --cluster CS6650L2-cluster --service-name CS6650L2 --region us-west-2

   # Get network interface
   aws ecs describe-tasks --cluster CS6650L2-cluster \
     --tasks <TASK_ARN> --region us-west-2 \
     --query 'tasks[0].attachments[0].details'

   # Get public IP
   aws ec2 describe-network-interfaces \
     --network-interface-ids <ENI_ID> --region us-west-2 \
     --query 'NetworkInterfaces[0].Association.PublicIp'
   ```

5. **Test the API:**
   ```bash
   curl http://54.202.185.46:8080/products/1
   ```

### Teardown

```bash
cd terraform
terraform destroy
```

## API Examples - All Response Codes

### POST /products/{productId}/details

**204 - Success (product created):**
```bash
curl -v -X POST http://54.202.185.46:8080/products/1/details \
  -H "Content-Type: application/json" \
  -d '{"product_id":1,"sku":"ABC-123","manufacturer":"Acme","category_id":10,"weight":500,"some_other_id":99}'

# Response: 204 No Content
```

**400 - Invalid input (missing required fields):**
```bash
curl -v -X POST http://54.202.185.46:8080/products/1/details \
  -H "Content-Type: application/json" \
  -d '{"product_id":1}'

# Response: 400
# {"error":"INVALID_INPUT","message":"Invalid input data","details":"Key: 'Product.SKU' Error:Field validation for 'SKU' failed on the 'required' tag..."}
```

**400 - Invalid product ID:**
```bash
curl -v -X POST http://54.202.185.46:8080/products/abc/details \
  -H "Content-Type: application/json" \
  -d '{"product_id":1,"sku":"ABC","manufacturer":"Acme","category_id":1,"weight":100,"some_other_id":1}'

# Response: 400
# {"error":"INVALID_INPUT","message":"Invalid product ID","details":"Product ID must be a positive integer"}
```

### GET /products/{productId}

**200 - Product found:**
```bash
curl -v http://54.202.185.46:8080/products/1

# Response: 200
# {"product_id":1,"sku":"ABC-123","manufacturer":"Acme","category_id":10,"weight":500,"some_other_id":99}
```

**404 - Product not found:**
```bash
curl -v http://54.202.185.46:8080/products/999

# Response: 404
# {"error":"NOT_FOUND","message":"Product not found","details":"No product found with ID 999"}
```

## Stress Testing with Locust

### Setup

```bash
pip install locust
```

### Run Tests

```bash
# Headless mode
locust -f locustfile.py --host=http://54.202.185.46:8080 \
  --users 50 --spawn-rate 5 --run-time 60s --headless --only-summary

# Or with Web UI
locust -f locustfile.py --host=http://54.202.185.46:8080
# Then open http://localhost:8089
```

### Test Results (50 users, 60 seconds)

| Metric | Value |
|--------|-------|
| Total Requests | 1,825 |
| Requests/sec | ~31 |
| Avg Response Time | 31ms |
| Min Response Time | 17ms |
| Max Response Time | 185ms |
| p50 | 30ms |
| p95 | 51ms |
| p99 | 69ms |
| Real Failures | 0 (all 404s are intentional test cases) |

### Test Design

The Locust test uses weighted tasks to simulate realistic e-commerce traffic:
- **GET existing product (weight 3):** Most common operation - customers browsing products
- **POST product details (weight 1):** Less frequent - admin adding/updating products
- **GET nonexistent product (weight 1):** Tests 404 handling under load

This 3:1:1 ratio reflects real-world patterns where reads far outnumber writes in e-commerce systems.

### HttpUser vs FastHttpUser

Both user types are included in the Locustfile. `FastHttpUser` uses the `geventhttpclient` library instead of Python's `requests` library, which reduces client-side overhead. At low concurrency (50 users) hitting a remote server, the difference is minimal because network latency (~25-30ms to AWS) dominates over client-side processing time. The difference becomes significant at high concurrency (1000+ users) where client-side overhead matters.

### Data Structure Choice

Products are stored in a Go `map[int]Product` protected by `sync.RWMutex`. This provides:
- **O(1) lookups** for GET requests (the most common operation)
- **O(1) inserts** for POST requests
- **Read-write lock** allows concurrent reads while serializing writes

For a read-heavy e-commerce workload, a hashmap is optimal. A slice would require O(n) lookups which degrades under load.

## Design Discussion

### Scalable Backend Design for Full API

To handle the complete e-commerce API (products, carts, warehouse, payments) at scale:

- **Microservices architecture:** Separate services for Products, Shopping Cart, Warehouse, and Payments, each independently scalable
- **API Gateway:** Single entry point routing to appropriate services
- **Database per service:** Products in a read-optimized store (Redis cache + PostgreSQL), Carts in Redis (ephemeral), Warehouse in PostgreSQL (ACID for inventory), Payments via external processor
- **Message queue (SQS/Kafka):** Async communication between checkout → warehouse → payments
- **Load balancer:** Distribute traffic across multiple ECS tasks per service
- **Auto-scaling:** ECS service auto-scaling based on CPU/request count

### Declarative vs Imperative

Terraform is **declarative**: you describe the *desired end state* ("I want an ECS cluster with 1 task running this image"), and Terraform figures out *how* to get there. This contrasts with **imperative** approaches (like a bash script) where you specify each step ("first create the cluster, then create the task definition, then...").

Benefits of declarative:
- **Idempotent:** Running `terraform apply` twice produces the same result
- **State tracking:** Terraform knows what exists and only changes what's needed
- **Drift detection:** Can detect and correct manual changes
- **Reproducible:** Same config always produces same infrastructure