from locust import HttpUser, FastHttpUser, task, between, events
import random
import string

class ProductUser(HttpUser):
    wait_time = between(1, 3)
    
    def on_start(self):
        """Seed some products on start"""
        for i in range(1, 11):
            self.client.post(f"/products/{i}/details", json={
                "product_id": i,
                "sku": f"SKU-{i:03d}",
                "manufacturer": f"Manufacturer-{i}",
                "category_id": random.randint(1, 50),
                "weight": random.randint(100, 5000),
                "some_other_id": random.randint(1, 100)
            })

    @task(3)
    def get_product(self):
        """GET is 3x more common than POST (realistic e-commerce)"""
        product_id = random.randint(1, 10)
        self.client.get(f"/products/{product_id}")

    @task(1)
    def add_product(self):
        product_id = random.randint(1, 100)
        self.client.post(f"/products/{product_id}/details", json={
            "product_id": product_id,
            "sku": f"SKU-{''.join(random.choices(string.ascii_uppercase, k=5))}",
            "manufacturer": "TestManufacturer",
            "category_id": random.randint(1, 50),
            "weight": random.randint(100, 5000),
            "some_other_id": random.randint(1, 100)
        })

    @task(1)
    def get_nonexistent(self):
        """Test 404 path"""
        self.client.get(f"/products/{random.randint(10000, 99999)}")


class FastProductUser(FastHttpUser):
    """Same tests but with FastHttpUser for comparison"""
    wait_time = between(1, 3)

    def on_start(self):
        for i in range(1, 11):
            self.client.post(f"/products/{i}/details", json={
                "product_id": i,
                "sku": f"SKU-{i:03d}",
                "manufacturer": f"Manufacturer-{i}",
                "category_id": random.randint(1, 50),
                "weight": random.randint(100, 5000),
                "some_other_id": random.randint(1, 100)
            })

    @task(3)
    def get_product(self):
        product_id = random.randint(1, 10)
        self.client.get(f"/products/{product_id}")

    @task(1)
    def add_product(self):
        product_id = random.randint(1, 100)
        self.client.post(f"/products/{product_id}/details", json={
            "product_id": product_id,
            "sku": f"SKU-{''.join(random.choices(string.ascii_uppercase, k=5))}",
            "manufacturer": "TestManufacturer",
            "category_id": random.randint(1, 50),
            "weight": random.randint(100, 5000),
            "some_other_id": random.randint(1, 100)
        })

    @task(1)
    def get_nonexistent(self):
        self.client.get(f"/products/{random.randint(10000, 99999)}")