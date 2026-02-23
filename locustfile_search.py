from locust import FastHttpUser, task, constant
import random

search_terms = ["alpha", "beta", "gamma", "electronics", "books", "home", "sports", "clothing"]

class SearchUser(FastHttpUser):
    wait_time = constant(0)  # No wait — maximum pressure

    @task
    def search_product(self):
        term = random.choice(search_terms)
        self.client.get(f"/products/search?q={term}")