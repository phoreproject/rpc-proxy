#!/usr/bin/python3.6

"""
Notifies the UNIX socket server about a new block.
"""

import socket
import sys


def main():
    "Sends a blockhash to the UNIX socket"
    blockhash = sys.argv[1]

    assert len(blockhash) == 64, "Hash must be exactly 64 hex characters long"

    blockhash_bytes = bytearray.fromhex(blockhash)

    sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)

    try:
        sock.connect('/tmp/blocknotify.sock')
    except socket.error as msg:
        print(msg)
        sys.exit(1)

    try:
        sock.sendall(blockhash_bytes)
    finally:
        sock.close()

if __name__ == '__main__':
    main()
