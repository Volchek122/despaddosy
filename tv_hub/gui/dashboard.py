from PyQt6.QtWidgets import QWidget, QVBoxLayout, QPushButton, QLabel, QGridLayout, QSizePolicy
from PyQt6.QtCore import pyqtSignal, Qt

class Dashboard(QWidget):
    app_launched = pyqtSignal(str)

    def __init__(self, parent=None):
        super().__init__(parent)
        self.layout = QVBoxLayout()
        self.setLayout(self.layout)

        # Title
        self.title_label = QLabel("TV Hub Dashboard")
        self.title_label.setAlignment(Qt.AlignmentFlag.AlignCenter)
        self.title_label.setStyleSheet("font-size: 32px; font-weight: bold; margin-bottom: 20px;")
        self.layout.addWidget(self.title_label)

        # Grid of apps
        self.grid_layout = QGridLayout()
        self.layout.addLayout(self.grid_layout)

        # Flixy Button
        self.flixy_btn = QPushButton("Flixy")
        self.flixy_btn.setSizePolicy(QSizePolicy.Policy.Expanding, QSizePolicy.Policy.Expanding)
        self.flixy_btn.setMinimumSize(200, 150)
        self.flixy_btn.setStyleSheet("""
            QPushButton {
                background-color: #E50914;
                color: white;
                border-radius: 10px;
                font-size: 24px;
                font-weight: bold;
            }
            QPushButton:hover {
                background-color: #f40612;
            }
        """)
        self.flixy_btn.clicked.connect(lambda: self.app_launched.emit("flixy"))
        self.grid_layout.addWidget(self.flixy_btn, 0, 0)

        # Exit Button
        self.exit_btn = QPushButton("Exit")
        self.exit_btn.setSizePolicy(QSizePolicy.Policy.Expanding, QSizePolicy.Policy.Expanding)
        self.exit_btn.setMinimumSize(200, 150)
        self.exit_btn.setStyleSheet("""
            QPushButton {
                background-color: #333333;
                color: white;
                border-radius: 10px;
                font-size: 24px;
            }
            QPushButton:hover {
                background-color: #444444;
            }
        """)
        self.exit_btn.clicked.connect(lambda: self.app_launched.emit("exit"))
        self.grid_layout.addWidget(self.exit_btn, 0, 1)

        # Add filler to push content to center properly if needed
        self.layout.addStretch()
