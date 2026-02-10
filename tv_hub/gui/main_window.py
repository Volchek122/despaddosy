from PyQt6.QtWidgets import (QMainWindow, QWidget, QVBoxLayout, QStackedWidget,
                             QLineEdit, QPushButton, QHBoxLayout, QLabel)
from PyQt6.QtCore import Qt
import sys
from tv_hub.gui.widgets.poster_grid import PosterGrid
from tv_hub.gui.widgets.details import DetailsView
from tv_hub.gui.widgets.player import PlayerView
from tv_hub.core.rezka import RezkaProvider
from tv_hub.core.mock import MockProvider

class MainWindow(QMainWindow):
    def __init__(self):
        super().__init__()
        self.setWindowTitle("Flixy - TV Hub")
        self.resize(1280, 720)

        # Determine provider (try Rezka, fallback to Mock)
        try:
            self.provider = RezkaProvider()
            # If no API key or failed init, maybe fallback
            if not getattr(self.provider, 'api', None) and not isinstance(self.provider, MockProvider):
                 # We can use MockProvider for specific features if Rezka fails hard,
                 # but RezkaProvider has its own fallback.
                 # Let's use MockProvider for now if RezkaProvider is just a shell.
                 pass
        except Exception:
            self.provider = MockProvider()

        # If RezkaProvider fails to load anything useful (e.g. cloudflare), we might want to toggle to MockProvider via UI
        # For now, let's just default to MockProvider if we detect issues, or user can switch.
        # Actually, let's use MockProvider by default for development stability unless configured otherwise.
        # Uncomment below to force real provider:
        # self.provider = RezkaProvider()
        self.provider = MockProvider() # Defaulting to Mock for reliable demo

        self.setup_ui()
        self.setup_connections()
        self.load_styles()

    def setup_ui(self):
        central_widget = QWidget()
        self.setCentralWidget(central_widget)
        layout = QVBoxLayout(central_widget)
        layout.setContentsMargins(0, 0, 0, 0)

        # Header / Search
        header = QWidget()
        header.setFixedHeight(60)
        header_layout = QHBoxLayout(header)

        self.search_input = QLineEdit()
        self.search_input.setPlaceholderText("Search movies, series...")
        self.search_input.setStyleSheet("padding: 10px; border-radius: 5px; font-size: 16px;")

        self.search_btn = QPushButton("Search")
        self.search_btn.setFixedWidth(100)
        self.search_btn.setStyleSheet("background-color: #e50914; color: white; border-radius: 5px; padding: 10px;")

        header_layout.addWidget(self.search_input)
        header_layout.addWidget(self.search_btn)

        layout.addWidget(header)

        # Main Content Area (Stacked)
        self.stack = QStackedWidget()

        # Page 1: Grid
        self.grid_page = QWidget()
        grid_layout = QVBoxLayout(self.grid_page)
        self.poster_grid = PosterGrid()
        grid_layout.addWidget(self.poster_grid)
        self.stack.addWidget(self.grid_page)

        # Page 2: Details
        self.details_page = DetailsView(self.provider)
        self.stack.addWidget(self.details_page)

        # Page 3: Player
        self.player_page = PlayerView()
        self.stack.addWidget(self.player_page)

        layout.addWidget(self.stack)

    def setup_connections(self):
        self.search_btn.clicked.connect(self.perform_search)
        self.search_input.returnPressed.connect(self.perform_search)
        self.poster_grid.item_clicked_signal.connect(self.show_details)
        self.details_page.back_signal.connect(self.show_grid)
        self.details_page.play_signal.connect(self.start_playback)
        self.player_page.close_signal.connect(self.stop_playback)

    def perform_search(self):
        query = self.search_input.text()
        if query:
            results = self.provider.search(query)
            self.poster_grid.add_items(results)
            self.stack.setCurrentWidget(self.grid_page)
            if results:
                self.poster_grid.setFocus()
                self.poster_grid.setCurrentRow(0)

    def show_details(self, item):
        # Fetch full details including seasons
        full_item = self.provider.get_details(item.url)
        self.details_page.set_item(full_item)
        self.stack.setCurrentWidget(self.details_page)

    def show_grid(self):
        self.stack.setCurrentWidget(self.grid_page)

    def start_playback(self, url):
        self.stack.setCurrentWidget(self.player_page)
        self.player_page.play(url)

    def stop_playback(self):
        self.stack.setCurrentWidget(self.details_page)

    def load_styles(self):
        with open("assets/style.qss", "r") as f:
            self.setStyleSheet(f.read())
