from PyQt6.QtWidgets import QMainWindow, QStackedWidget, QWidget, QVBoxLayout, QApplication
from PyQt6.QtCore import Qt, QSize, QPoint, QEvent
from PyQt6.QtGui import QCursor, QMouseEvent
from .dashboard import Dashboard
from .flixy_app import FlixyApp
from .remote_listener import RemoteListener

class MainWindow(QMainWindow):
    def __init__(self):
        super().__init__()
        self.setWindowTitle("TV Hub")
        self.resize(1280, 720)

        # Style
        self.setStyleSheet("""
            QMainWindow {
                background-color: #121212;
                color: #ffffff;
            }
            QWidget {
                font-family: Arial, sans-serif;
                font-size: 18px;
            }
        """)

        self.central_widget = QStackedWidget()
        self.setCentralWidget(self.central_widget)

        self.dashboard = Dashboard(self)
        self.central_widget.addWidget(self.dashboard)

        # Connect signals
        self.dashboard.app_launched.connect(self.launch_app)

        # Remote Control
        self.remote_listener = RemoteListener()
        self.remote_listener.command_received.connect(self.handle_remote_command)
        self.remote_listener.start()

    def handle_remote_command(self, cmd):
        command_type = cmd.get('cmd')
        if command_type == 'move':
            dx = cmd.get('dx', 0)
            dy = cmd.get('dy', 0)
            current_pos = QCursor.pos()
            new_pos = current_pos + QPoint(int(dx), int(dy))
            # Ensure within screen bounds? QCursor handles it mostly, but good to clamp if needed.
            QCursor.setPos(new_pos)

        elif command_type == 'click':
            # Simulate click at current position
            pos = QCursor.pos()
            # Find widget at position
            widget = QApplication.widgetAt(pos)
            if widget:
                # Map global pos to widget local pos
                local_pos = widget.mapFromGlobal(pos)

                # Press
                event_press = QMouseEvent(QEvent.Type.MouseButtonPress, local_pos,
                                          Qt.MouseButton.LeftButton, Qt.MouseButton.LeftButton, Qt.KeyboardModifier.NoModifier)
                QApplication.postEvent(widget, event_press)

                # Release
                event_release = QMouseEvent(QEvent.Type.MouseButtonRelease, local_pos,
                                            Qt.MouseButton.LeftButton, Qt.MouseButton.LeftButton, Qt.KeyboardModifier.NoModifier)
                QApplication.postEvent(widget, event_release)

        elif command_type == 'back':
            # Simulate Escape key
            from PyQt6.QtGui import QKeyEvent
            event = QKeyEvent(QEvent.Type.KeyPress, Qt.Key.Key_Escape, Qt.KeyboardModifier.NoModifier)
            QApplication.postEvent(self, event)

    def closeEvent(self, event):
        self.remote_listener.stop()
        super().closeEvent(event)

    def launch_app(self, app_name):
        if app_name == "flixy":
            print("Launching Flixy...")
            self.flixy_app = FlixyApp()
            self.flixy_app.back_to_dashboard.connect(self.back_to_dashboard)
            self.central_widget.addWidget(self.flixy_app)
            self.central_widget.setCurrentWidget(self.flixy_app)
        elif app_name == "exit":
            self.close()

    def back_to_dashboard(self):
        # Remove Flixy and go back to Dashboard
        self.central_widget.setCurrentIndex(0) # Dashboard is at 0
        if hasattr(self, 'flixy_app'):
            self.central_widget.removeWidget(self.flixy_app)
            self.flixy_app.deleteLater()
            del self.flixy_app

    def keyPressEvent(self, event):
        if event.key() == Qt.Key.Key_Escape:
            current_widget = self.central_widget.currentWidget()
            if isinstance(current_widget, FlixyApp):
                current_widget.handle_back()
            elif self.central_widget.currentIndex() > 0:
                self.central_widget.setCurrentIndex(0) # Go back to dashboard
            else:
                self.close()
