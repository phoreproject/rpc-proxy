"""
Notification listener listens to a UNIX socket in the same directory
and relays block notifications to a websocket server.

Author: Julian Meyer <meyer9>
"""

import json
import socket
import os
import binascii
import threading

import asyncio
import websockets

class EventThreadSafe(asyncio.Event):
    "Thread-safe async event"
    def set(self):
        self._loop.call_soon_threadsafe(super().set)

async def websocket_thread(new_block: EventThreadSafe, block_data):
    "Runs a websocket server and relays events from UNIX socket thread."
    subscribed = []

    async def wait_for_event(event: EventThreadSafe):
        "Waits for the event to become set asynchronously."
        await event.wait()

    async def notify_clients():
        "Notifies clients of a new block hash"
        nonlocal block_data
        while True:
            await wait_for_event(new_block)

            for subscriber in subscribed:
                await subscriber(block_data[0])

            block_data.clear()
            new_block.clear()

    async def handle_connection(websocket, path):
        "Handles a single connection to websockets."
        async for message in websocket:
            if message == "subscribe":
                subscribed.append(websocket.send)
                await websocket.send(json.dumps({"status": "success"}))

    await websockets.serve(handle_connection, 'localhost', 8765)
    await notify_clients()


def block_notify_listener_thread(block_notify_event: EventThreadSafe, block_data):
    "Runs a UNIX socker server and relays events to the websocket thread."
    server_address = '/tmp/blocknotify.sock'

    # Make sure the socket does not already exist
    try:
        os.unlink(server_address)
    except OSError:
        if os.path.exists(server_address):
            raise

    sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)

    sock.bind(server_address)

    sock.listen(0)

    while True:
        connection, _ = sock.accept()
        try:
            while True:
                data = connection.recv(32)
                if data:
                    print('received {}'.format(binascii.hexlify(data)))
                    block_data.append(data)
                    block_notify_event.set()
                else:
                    break
        finally:
            # Clean up the connection
            connection.close()

async def main():
    new_block = EventThreadSafe()
    block_data = []
    print("[+] Starting block notify listener thread")
    block_notify_thread = threading.Thread(
        target=block_notify_listener_thread, args=(new_block, block_data,))
    block_notify_thread.start()
    await websocket_thread(new_block, block_data)

if __name__ == '__main__':
    asyncio.ensure_future(main())
    asyncio.get_event_loop().run_forever()
