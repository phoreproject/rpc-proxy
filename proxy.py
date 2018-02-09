"""
A proxy on top of JSON-RPC to interact with the Phore network without
syncing to the network or running a full node.

Author: Julian Meyer <meyer9>
"""

from http.server import BaseHTTPRequestHandler, HTTPServer
import requests
import json

# These commands should not carry state, require sensitive information, or send sensitive data.
ALLOWED_COMMANDS = [
    'getbestblockhash',
    'getblock',
    'getblockchaininfo',
    'getblockcount',
    'getblockhash',
    'getblockheader',
    'getchaintips',
    'getdifficulty',
    'getmempoolinfo',
    'getrawmempool',
    'gettxout',
    'gettxoutsetinfo',
    'getinfo',
    'getmininginfo',
    'getnetworkhashps',
    'submitblock',
    'getconnectioncount',
    'ping',
    'masternodelist',
    'getrawtransaction',
    'sendrawtransaction',
    'estimatefee',
    'estimatepriority',
    'searchrawtransactions'
]

def rpc_proxy_class(config_data):
    """
    Generates an RPCProxy request handler provided a given
    config file.
    """

    rpcuri = 'http://{}:{}'.format(config_data.get('rpchost'), config_data.get('rpcport'))

    class RpcProxy(BaseHTTPRequestHandler):
        """
        Passes through data to JSON-RPC provided it is
        JSON decodable and allowed (stateless and unpriveleged)
        """

        def respond(self, data):
            """
            Sends data and a 200 OK response.
            """
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.send_header('Content-Length', str(len(data)))
            self.send_header('Access-Control-Allow-Origin', '*')
            self.end_headers()
            self.wfile.write(data)

        def respond_with_json(self, obj):
            """
            Sends JSON and a 200 OK repsonse.
            """
            out = json.dumps(obj).encode()
            self.respond(out)

        def bad_request(self, err):
            """
            Sends a bad request response to the client.
            """
            self.send_response(400)
            self.send_header('Content-type', 'application/json')
            self.respond_with_json({'err': err})

        def forbidden(self, cmd):
            """
            Sends a forbidden response to the client.
            """
            self.send_response(403)
            self.send_header('Content-type', 'application/json')
            self.respond_with_json({'err': 'Forbidden to run command "{}"'.format(cmd)})

        def do_POST(self):
            """
            Handles a POST request
            """
            length = int(self.headers.get('content-length'))
            raw_data = self.rfile.read(length)
            try:
                json_data = json.loads(raw_data)
            except json.JSONDecodeError:
                return self.bad_request('could not decode json')
            if json_data.get('method') not in ALLOWED_COMMANDS:
                return self.forbidden(json_data.get('method'))
            response = requests.get(rpcuri, headers={'content-type': 'application/json'}, data=raw_data, auth=(config_data.get('rpcusername'), config_data.get('rpcpassword')))

            self.respond_with_json(response.json())
    return RpcProxy


def run(config_data):
    """
    Runs the server given a config file.
    """
    server_address = (config_data['host'], config_data['port'])
    httpd = HTTPServer(server_address, rpc_proxy_class(config_data))
    print('Starting rpc proxy...')
    print('Running on %s:%s' % (server_address[0], server_address[1]))
    httpd.serve_forever()

if __name__ == "__main__":
    with open('config.json', 'r') as config_file:
        config_data = json.load(config_file)
        run(config_data)
