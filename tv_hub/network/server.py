import socket
import threading
import json
import time

class DiscoveryServer(threading.Thread):
    def __init__(self, port=5000):
        super().__init__()
        self.port = port
        self.running = True
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self.sock.bind(('0.0.0.0', self.port))
        print(f"[Discovery] Listening on UDP {self.port}")

    def run(self):
        while self.running:
            try:
                data, addr = self.sock.recvfrom(1024)
                message = data.decode('utf-8').strip()
                if message == "DISCOVER_TV_HUB":
                    # Reply to the client
                    reply = "TV_HUB_HERE"
                    self.sock.sendto(reply.encode('utf-8'), addr)
                    print(f"[Discovery] Responded to {addr}")
            except Exception as e:
                if self.running:
                    print(f"[Discovery] Error: {e}")

    def stop(self):
        self.running = False
        self.sock.close()

class ControlServer(threading.Thread):
    def __init__(self, host='0.0.0.0', port=5001, command_callback=None):
        super().__init__()
        self.host = host
        self.port = port
        self.running = True
        self.command_callback = command_callback
        self.server_sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.server_sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self.server_sock.bind((self.host, self.port))
        self.server_sock.listen(5)
        print(f"[Control] Listening on TCP {self.port}")

    def run(self):
        while self.running:
            try:
                client_sock, addr = self.server_sock.accept()
                print(f"[Control] Connected by {addr}")
                client_handler = threading.Thread(
                    target=self.handle_client,
                    args=(client_sock,)
                )
                client_handler.start()
            except OSError:
                break # Socket closed
            except Exception as e:
                print(f"[Control] Accept Error: {e}")

    def handle_client(self, sock):
        buffer = ""
        with sock:
            while True:
                try:
                    data = sock.recv(4096)
                    if not data:
                        break
                    buffer += data.decode('utf-8')

                    # Simple JSON framing: Assuming newline delimited JSON for simplicity
                    while '\n' in buffer:
                        message, buffer = buffer.split('\n', 1)
                        if message.strip():
                            self.process_command(message)
                except Exception as e:
                    print(f"[Control] Client Error: {e}")
                    break
        print("[Control] Client disconnected")

    def process_command(self, message_str):
        try:
            command = json.loads(message_str)
            if self.command_callback:
                self.command_callback(command)
            else:
                print(f"[Control] Received: {command}")
        except json.JSONDecodeError:
            print(f"[Control] Invalid JSON: {message_str}")

    def stop(self):
        self.running = False
        self.server_sock.close()

if __name__ == "__main__":
    # Test execution
    disc = DiscoveryServer()
    ctrl = ControlServer()

    disc.start()
    ctrl.start()

    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        print("Stopping servers...")
        disc.stop()
        ctrl.stop()
