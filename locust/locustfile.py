import random
from locust import HttpUser, FastHttpUser, between

NUM_PRODUCTS = 100  # matches NewProductStore(100) in the server


def get_product(user):
    pid = random.randint(1, NUM_PRODUCTS)
    user.client.get(f"/products/{pid}", name="GET /products/[id]")


def update_product(user):
    pid = random.randint(1, NUM_PRODUCTS)
    body = {
        "product_id": pid,
        "sku": f"SKU-{pid}",
        "manufacturer": "Acme Corporation",
        "category_id": 1,
        "weight": random.randint(0, 5000),
        "some_other_id": 1,
    }
    user.client.post(f"/products/{pid}/details", json=body, name="POST /products/[id]/details")


# Request mix: 90% reads, 10% writes
TASKS = {get_product: 9, update_product: 1}


class ProductHttpUser(HttpUser):
    tasks = TASKS
    wait_time = between(0.1, 0.5)


class ProductFastHttpUser(FastHttpUser):
    tasks = TASKS
    wait_time = between(0.1, 0.5)