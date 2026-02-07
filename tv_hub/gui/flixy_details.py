from PyQt6.QtWidgets import (QWidget, QVBoxLayout, QLabel, QPushButton,
                             QHBoxLayout, QComboBox, QScrollArea, QSizePolicy)
from PyQt6.QtCore import Qt, pyqtSignal
from PyQt6.QtGui import QPixmap
import requests
from tv_hub.gui.player import VideoPlayer

class FlixyDetails(QWidget):
    back_clicked = pyqtSignal()
    play_clicked = pyqtSignal(str) # Emits stream URL

    def __init__(self, movie_data, provider, parent=None):
        super().__init__(parent)
        self.movie_data = movie_data
        self.provider = provider

        self.layout = QVBoxLayout()
        self.setLayout(self.layout)

        # Back Button
        self.back_btn = QPushButton("← Back")
        self.back_btn.setStyleSheet("font-size: 18px; padding: 10px; background-color: #333; color: white;")
        self.back_btn.clicked.connect(self.back_clicked.emit)
        self.layout.addWidget(self.back_btn, alignment=Qt.AlignmentFlag.AlignLeft)

        # Content Layout
        content_layout = QHBoxLayout()
        self.layout.addLayout(content_layout)

        # Poster (Placeholder for now, could reuse AsyncImageLoader)
        self.poster_lbl = QLabel()
        self.poster_lbl.setFixedSize(300, 450)
        self.poster_lbl.setStyleSheet("background-color: #222;")
        self.poster_lbl.setAlignment(Qt.AlignmentFlag.AlignCenter)
        if movie_data.get('poster_url'):
            # Ideally load async, for prototype load sync or just text
            self.poster_lbl.setText("Loading Image...")
            # In real app, re-use the cached pixmap or async loader
        else:
             self.poster_lbl.setText("No Image")
        content_layout.addWidget(self.poster_lbl)

        # Info
        info_layout = QVBoxLayout()
        content_layout.addLayout(info_layout)

        title = QLabel(f"{movie_data['title']} ({movie_data['year']})")
        title.setStyleSheet("font-size: 32px; font-weight: bold; margin-bottom: 10px;")
        info_layout.addWidget(title)

        desc = QLabel(movie_data.get('description', 'No description available.'))
        desc.setWordWrap(True)
        desc.setStyleSheet("font-size: 18px; color: #ccc; margin-bottom: 20px;")
        info_layout.addWidget(desc)

        # Season/Episode Selector
        self.season_combo = QComboBox()
        self.episode_combo = QComboBox()

        if movie_data.get('seasons'):
            self.season_lbl = QLabel("Season:")
            info_layout.addWidget(self.season_lbl)
            self.season_combo.addItems([str(s) for s in movie_data['seasons']])
            self.season_combo.currentIndexChanged.connect(self.update_episodes)
            info_layout.addWidget(self.season_combo)

            self.episode_lbl = QLabel("Episode:")
            info_layout.addWidget(self.episode_lbl)
            info_layout.addWidget(self.episode_combo)
            self.update_episodes() # Init episodes
        else:
            self.season_combo.hide()
            self.episode_combo.hide()

        # Play Button
        self.play_btn = QPushButton("▶ Play")
        self.play_btn.setMinimumHeight(50)
        self.play_btn.setStyleSheet("""
            QPushButton {
                background-color: #E50914;
                color: white;
                font-size: 24px;
                font-weight: bold;
                border-radius: 5px;
            }
            QPushButton:hover {
                background-color: #f40612;
            }
        """)
        self.play_btn.clicked.connect(self.on_play)
        info_layout.addWidget(self.play_btn)

        info_layout.addStretch()

    def update_episodes(self):
        season = int(self.season_combo.currentText())
        episodes_count = self.movie_data['episodes'].get(season, 0)
        self.episode_combo.clear()
        self.episode_combo.addItems([str(i) for i in range(1, episodes_count + 1)])

    def on_play(self):
        season = None
        episode = None
        if self.season_combo.isVisible():
            season = int(self.season_combo.currentText())
            episode = int(self.episode_combo.currentText())

        # Get stream URL from provider
        url = self.provider.get_stream_url(self.movie_data['id'], season, episode)
        self.play_clicked.emit(url)
