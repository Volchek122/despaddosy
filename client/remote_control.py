import socket
import threading
import json
import time

class RemoteControl:
    def __init__(self, port=5000):
        self.port = port
        self.host = None
        self.sock = None
        self.running = True

    def discover_hub(self):
        print("[Client] Searching for TV Hub...")
        udp_sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        udp_sock.setsockopt(socket.SOL_SOCKET, socket.SO_BROADCAST, 1)
        udp_sock.settimeout(5.0)

        udp_sock.bind(('', 0))  # Bind to any available port for receiving reply

        try:
            message = "DISCOVER_TV_HUB"
            udp_sock.sendto(message.encode('utf-8'), ('255.255.255.255', self.port))

            # Listen for reply
            data, addr = udp_sock.recvfrom(1024)
            reply = data.decode('utf-8').strip()

            if reply == "TV_HUB_HERE":
                print(f"[Client] Found Hub at {addr}")
                self.host = addr[0]
                return True
        except socket.timeout:
            print("[Client] Discovery timed out. No Hub found.")
            return False
        except Exception as e:
            print(f"[Client] Discovery Error: {e}")
            return False
        finally:
            udp_sock.close()

    def connect(self):
        if not self.host:
            print("[Client] No host found via discovery.")
            return False

        try:
            self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.sock.connect((self.host, 5001))
            print(f"[Client] Connected to {self.host}:5001")
            return True
        except Exception as e:
            print(f"[Client] Connection Error: {e}")
            return False

    def send_command(self, command):
        if not self.sock:
            print("[Client] Not connected.")
            return

        try:
            # Send newline delimited JSON
            message = json.dumps(command) + "\n"
            self.sock.sendall(message.encode('utf-8'))
            print(f"[Client] Sent: {command}")
        except Exception as e:
            print(f"[Client] Send Error: {e}")
            self.sock.close()
            self.sock = None

    def simulate_gyro(self):
        """Simulates gyro movement by sending small deltas."""
        print("[Client] Simulating Gyroscope...")
        try:
            for i in range(5):
                # Send a 'move' command
                self.send_command({"cmd": "move", "dx": 5, "dy": 0})
                time.sleep(0.5)
            # Send a 'click' command
            self.send_command({"cmd": "click", "button": "left"})
        except KeyboardInterrupt:
            pass

    def close(self):
        if self.sock:
            self.sock.close()
            print("[Client] Disconnected.")

if __name__ == "__main__":
    remote = RemoteControl()
    if remote.discover_hub():
        if remote.connect():
            remote.simulate_gyro()
            remote.close()
