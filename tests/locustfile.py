import sys
import os
import json
import uuid
import time
import string
import random
import requests

import gevent
from websocket import create_connection
from locust import HttpUser, TaskSet, task, between, events

# HOSTNAME = "localhost:8000"
HOSTNAME = "nime2020.rytrose.com"

@events.init.add_listener
def init_locust(environment, **kwargs):
    print("Turning off websocket CORS")
    admin_key = os.environ.get("ADMIN_KEY")
    if admin_key == None or admin_key == "":
        sys.exit("No admin key found in environment (ADMIN_KEY)")
    res = requests.post(f"http://{HOSTNAME}/admin/websocket/cors?enforce=false", headers={
        "X-Admin-Key": admin_key
    })
    if res.status_code != 202:
        sys.exit("Unable to turn off websocket CORS enforcement")


@events.quitting.add_listener
def quit_locust(environment):
    print("Turning on websocket CORS")
    admin_key = os.environ.get("ADMIN_KEY")
    if admin_key == None or admin_key == "":
        sys.exit("No admin key found in environment (ADMIN_KEY)")
    res = requests.post(f"http://{HOSTNAME}/admin/websocket/cors?enforce=true", headers={
        "X-Admin-Key": admin_key
    })
    if res.status_code != 202:
        sys.exit("Unable to turn on websocket CORS enforcement")


def random_string(length=10):
    characters = string.ascii_letters + string.digits + string.punctuation
    return ''.join(random.choice(characters) for i in range(length))

class CommitOperationTaskSet(TaskSet):
    def on_start(self):
        self.id = str(uuid.uuid4())
        self.ws = create_connection(f"ws://{HOSTNAME}/ws")

        # Announce user
        announce_msg = json.dumps({
            "type": "announce",
            "userID": self.id
        })
        self.ws.send(announce_msg)

        # Enter room "test"
        enter_room_msg = json.dumps({
            "type": "enterRoom",
            "roomName": "test"
        })
        self.ws.send(enter_room_msg)

        def _receive():
            while True:
                res = self.ws.recv()
                data = json.loads(res)
                if data.get("type") == "operationsUpdate":
                    response_time = (time.time() - data['messageTime']) * 1000
                    events.request_success.fire(
                        request_type='Operations Receive',
                        name='operations/ws/recv',
                        response_time=response_time,
                        response_length=len(res),
                    )

        gevent.spawn(_receive)

    def on_quit(self):
        self.ws.close()

    @task
    def sent(self):
        message_time = time.time()
        msg = json.dumps({
            "type": "operations",
            "operations": [
                {
                    "randomData": random_string(random.randint(50, 200))
                }
                for _ in range(random.randint(1, 10))
            ],
            "messageTime": time.time()
        })
        self.ws.send(msg)
        events.request_success.fire(
            request_type='Operations Sent',
            name='operations/ws/send',
            response_time=(time.time() - message_time) * 1000,
            response_length=len(msg),
        )


class OperationsLocust(HttpUser):
    tasks = [CommitOperationTaskSet]
    wait_time = between(1, 2)
