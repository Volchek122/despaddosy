import requests
from PyQt6.QtWidgets import (QWidget, QVBoxLayout, QListWidget, QListWidgetItem,
                             QLabel, QAbstractItemView, QHBoxLayout, QLineEdit, QPushButton)
from PyQt6.QtCore import Qt, QThread, pyqtSignal, QSize
from PyQt6.QtGui import QIcon, QPixmap

from tv_hub.content.provider import MockProvider
# In a real app, we might inject the provider. For now, we instantiate it here or pass it in.

class ImageLoader(QThread):
    image_loaded = pyqtSignal(str, QPixmap)

    def __init__(self, url, item_id):
        super().__init__()
        self.url = url
        self.item_id = item_id

    def run(self):
        try:
            if not self.url:
                return
            response = requests.get(self.url, timeout=10)
            response.raise_for_status()
            pixmap = QPixmap()
            pixmap.loadFromData(response.content)
            self.image_loaded.emit(self.item_id, pixmap)
        except Exception as e:
            print(f"Image load error for {self.url}: {e}")

class FlixyBrowser(QWidget):
    item_clicked = pyqtSignal(object) # Emits the movie data dict

    def __init__(self, provider=None, parent=None):
        super().__init__(parent)
        self.provider = provider if provider else MockProvider()
        self.layout = QVBoxLayout()
        self.setLayout(self.layout)

        # Title / Header / Search
        header_layout = QHBoxLayout()
        self.layout.addLayout(header_layout)

        self.header = QLabel("Flixy - Popular")
        self.header.setStyleSheet("font-size: 24px; font-weight: bold;")
        header_layout.addWidget(self.header)

        header_layout.addStretch()

        self.search_input = QLineEdit()
        self.search_input.setPlaceholderText("Search...")
        self.search_input.setStyleSheet("padding: 5px; border-radius: 5px; color: black; background-color: #ddd;")
        self.search_input.setFixedWidth(250)
        self.search_input.returnPressed.connect(self.perform_search)
        header_layout.addWidget(self.search_input)

        self.search_btn = QPushButton("🔍")
        self.search_btn.setFixedWidth(40)
        self.search_btn.setStyleSheet("background-color: #E50914; color: white; font-weight: bold; border-radius: 5px;")
        self.search_btn.clicked.connect(self.perform_search)
        header_layout.addWidget(self.search_btn)

        # Grid View
        self.list_widget = QListWidget()
        self.list_widget.setViewMode(QListWidget.ViewMode.IconMode)
        self.list_widget.setIconSize(QSize(150, 220))
        self.list_widget.setResizeMode(QListWidget.ResizeMode.Adjust)
        self.list_widget.setSpacing(20)
        self.list_widget.setMovement(QListWidget.Movement.Static)
        self.list_widget.setSelectionMode(QAbstractItemView.SelectionMode.SingleSelection)
        self.list_widget.setStyleSheet("""
            QListWidget {
                background-color: #121212;
                border: none;
            }
            QListWidget::item {
                color: #ffffff;
                border-radius: 5px;
                padding: 10px;
            }
            QListWidget::item:selected {
                background-color: #333333;
                border: 2px solid #E50914;
            }
            QListWidget::item:hover {
                background-color: #222222;
            }
        """)
        self.list_widget.itemClicked.connect(self.on_item_clicked)
        self.layout.addWidget(self.list_widget)

        self.load_content()
        self.loaders = [] # Keep references to prevent garbage collection

    def load_content(self, query=None):
        self.list_widget.clear()
        self.loaders.clear() # Stop existing loaders? Ideally yes, but tricky with threads.
        # Python GC might handle them if I clear the list, but threads keep running until done.

        if query:
            self.header.setText(f"Search Results: {query}")
            movies = self.provider.search(query)
        else:
            self.header.setText("Flixy - Popular")
            movies = self.provider.get_popular()

        if not movies and query:
            self.header.setText(f"No results for: {query}")

        for movie in movies:
            item = QListWidgetItem(movie['title'])
            # Store movie data in the item
            item.setData(Qt.ItemDataRole.UserRole, movie)

            # Set placeholder icon
            placeholder = QPixmap(150, 220)
            placeholder.fill(Qt.GlobalColor.gray)
            item.setIcon(QIcon(placeholder))

            self.list_widget.addItem(item)

            # Start async load
            if movie.get('poster_url'):
                loader = ImageLoader(movie['poster_url'], movie['id'])
                loader.image_loaded.connect(self.on_image_loaded)
                self.loaders.append(loader)
                loader.start()

    def perform_search(self):
        query = self.search_input.text().strip()
        if query:
            self.load_content(query)
        else:
            self.load_content() # Reload popular

    def on_image_loaded(self, item_id, pixmap):
        # Find item by ID
        for i in range(self.list_widget.count()):
            item = self.list_widget.item(i)
            data = item.data(Qt.ItemDataRole.UserRole)
            if data['id'] == item_id:
                item.setIcon(QIcon(pixmap))
                break

    def on_item_clicked(self, item):
        movie_data = item.data(Qt.ItemDataRole.UserRole)
        self.item_clicked.emit(movie_data)
