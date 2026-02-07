from PyQt6.QtCore import QObject, pyqtSignal
from tv_hub.network.server import DiscoveryServer, ControlServer

class RemoteListener(QObject):
    command_received = pyqtSignal(dict)

    def __init__(self):
        super().__init__()
        self.discovery = DiscoveryServer()
        self.control = ControlServer(command_callback=self.on_command)

    def start(self):
        self.discovery.start()
        self.control.start()

    def stop(self):
        self.discovery.stop()
        self.control.stop()

    def on_command(self, command):
        # This runs in the server thread. Emitting signal is thread-safe.
        self.command_received.emit(command)
