from http.server import BaseHTTPRequestHandler, HTTPServer

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

def run_server():
    server_address = ('', 8080)
    httpd = HTTPServer(server_address, MyHTTPRequestHandler)
    print('Server running on port 8080')
    httpd.serve_forever()

run_server()