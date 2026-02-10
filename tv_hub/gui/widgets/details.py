from PyQt6.QtWidgets import (QWidget, QVBoxLayout, QHBoxLayout, QLabel,
                             QPushButton, QComboBox, QScrollArea, QListWidget)
from PyQt6.QtCore import Qt, pyqtSignal
from PyQt6.QtGui import QFont

class DetailsView(QWidget):
    back_signal = pyqtSignal()
    play_signal = pyqtSignal(str) # URL

    def __init__(self, provider):
        super().__init__()
        self.provider = provider
        self.item = None

        self.layout = QVBoxLayout(self)
        self.layout.setContentsMargins(20, 20, 20, 20)

        # Header (Back button)
        self.header = QWidget()
        header_layout = QHBoxLayout(self.header)
        self.back_btn = QPushButton("← Back")
        self.back_btn.setObjectName("back_btn")
        self.back_btn.setFixedSize(100, 40)
        # self.back_btn.setStyleSheet("background-color: transparent; color: white; border: 1px solid white; border-radius: 5px;") # Moved to QSS
        self.back_btn.clicked.connect(self.back_signal.emit)
        header_layout.addWidget(self.back_btn)
        header_layout.addStretch()
        self.layout.addWidget(self.header)

        # Content
        content_layout = QHBoxLayout()

        # Poster (Left)
        self.poster_label = QLabel()
        self.poster_label.setFixedSize(300, 450)
        self.poster_label.setStyleSheet("background-color: #333; border-radius: 10px;")
        self.poster_label.setAlignment(Qt.AlignmentFlag.AlignCenter)
        content_layout.addWidget(self.poster_label)

        # Details (Right)
        details_layout = QVBoxLayout()
        self.title_label = QLabel("Title")
        self.title_label.setFont(QFont("Arial", 24, QFont.Weight.Bold))
        self.title_label.setStyleSheet("color: white;")

        self.meta_label = QLabel("2023 | TV-MA | Series")
        self.meta_label.setStyleSheet("color: #ccc; font-size: 14px;")

        self.desc_label = QLabel("Description...")
        self.desc_label.setWordWrap(True)
        self.desc_label.setStyleSheet("color: #ddd; font-size: 16px; margin-top: 10px;")

        # Season Selector
        self.season_combo = QComboBox()
        self.season_combo.setStyleSheet("background-color: #333; color: white; padding: 5px;")
        self.season_combo.currentIndexChanged.connect(self.load_episodes)

        # Episode List
        self.episode_list = QListWidget()
        self.episode_list.setObjectName("episode_list")
        # self.episode_list.setStyleSheet("background-color: transparent; color: white; border: none;") # Moved to QSS
        self.episode_list.itemClicked.connect(self.play_episode)

        details_layout.addWidget(self.title_label)
        details_layout.addWidget(self.meta_label)
        details_layout.addWidget(self.desc_label)
        details_layout.addWidget(self.season_combo)
        details_layout.addWidget(self.episode_list)
        details_layout.addStretch()

        content_layout.addLayout(details_layout)
        self.layout.addLayout(content_layout)

    def set_item(self, item):
        self.item = item
        self.title_label.setText(item.title)
        self.meta_label.setText(f"{item.year} | {item.rating} | {item.type.title()}")
        self.desc_label.setText(item.description or "No description available.")

        # Poster (placeholder for now)
        self.poster_label.setText(item.title[:2])

        # Seasons
        self.season_combo.clear()
        if item.seasons:
            for season in item.seasons:
                self.season_combo.addItem(season.title, season.id)
            self.load_episodes(0) # Load first season
        else:
            # If movie, just show "Play Movie"
            self.episode_list.clear()
            self.episode_list.addItem("Play Movie")

    def load_episodes(self, index):
        if not self.item or not self.item.seasons:
            return

        season_id = self.season_combo.itemData(index)
        # In a real app, we might need to fetch episodes for this season ID if not already loaded
        # But our model has them pre-loaded for MockProvider
        season = next((s for s in self.item.seasons if s.id == season_id), None)

        self.episode_list.clear()
        from PyQt6.QtWidgets import QListWidgetItem

        if season:
            for ep in season.episodes:
                list_item = QListWidgetItem(ep.title)
                list_item.setData(Qt.ItemDataRole.UserRole, ep)
                self.episode_list.addItem(list_item)

    def play_episode(self, list_item):
        ep = list_item.data(Qt.ItemDataRole.UserRole)
        if not ep:
            # Maybe it's a movie
            if self.item.type == 'movie':
                stream_url = self.provider.get_stream(self.item.url)
            else:
                return
        else:
            stream_url = self.provider.get_stream(ep.url)

        print(f"Playing stream: {stream_url}")
        self.play_signal.emit(stream_url)
