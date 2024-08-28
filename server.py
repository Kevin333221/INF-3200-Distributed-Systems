from http.server import BaseHTTPRequestHandler, HTTPServer
import sys

class MyHTTPRequestHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/helloworld':
            self.send_response(200)
            self.send_header('Content-type', 'text/plain')
            self.end_headers()
            host_port = f"{self.server.server_name}:{self.server.server_port}"
            self.wfile.write(host_port.encode())
        else:
            self.send_response(404)
            self.send_header('Content-type', 'text/plain')
            self.end_headers()
            self.wfile.write(b'404 Not Found')

def run_server(port):
    server_address = ('', port)
    httpd = HTTPServer(server_address, MyHTTPRequestHandler)
    print(f'Server running on port {port}')
    httpd.serve_forever()


if __name__ == '__main__':

    port = sys.argv[1]
    print(f"Port: {port}")
    run_server(int(port))