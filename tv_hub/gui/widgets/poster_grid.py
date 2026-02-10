from PyQt6.QtWidgets import QListWidget, QListWidgetItem, QAbstractItemView, QWidget, QVBoxLayout, QLabel
from PyQt6.QtCore import Qt, QSize, pyqtSignal
from PyQt6.QtGui import QIcon, QPixmap
import requests

class PosterGrid(QListWidget):
    item_clicked_signal = pyqtSignal(object)  # Emits the ContentItem

    def __init__(self):
        super().__init__()
        self.setViewMode(QListWidget.ViewMode.IconMode)
        self.setResizeMode(QListWidget.ResizeMode.Adjust)
        self.setSpacing(15)
        self.setIconSize(QSize(180, 270))
        self.setMovement(QListWidget.Movement.Static)
        self.setSelectionMode(QAbstractItemView.SelectionMode.SingleSelection)
        self.setVerticalScrollMode(QAbstractItemView.ScrollMode.ScrollPerPixel)
        # self.setStyleSheet("background-color: transparent; border: none;") # Moved to QSS
        self.itemClicked.connect(self._on_item_clicked)

    def add_items(self, items):
        self.clear()
        for item in items:
            list_item = QListWidgetItem(item.title)
            # In a real app, load images asynchronously. Here we just set a placeholder or load if possible.
            # For now, we'll use a placeholder icon.
            icon = QIcon("assets/placeholder_poster.png")
            # If we want to load the real image:
            # pixmap = QPixmap()
            # data = requests.get(item.poster_url).content
            # pixmap.loadFromData(data)
            # icon = QIcon(pixmap)

            list_item.setIcon(icon)
            list_item.setData(Qt.ItemDataRole.UserRole, item)
            self.addItem(list_item)

    def _on_item_clicked(self, item):
        content_item = item.data(Qt.ItemDataRole.UserRole)
        self.item_clicked_signal.emit(content_item)
