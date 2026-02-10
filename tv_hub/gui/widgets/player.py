from PyQt6.QtWidgets import QWidget, QVBoxLayout, QPushButton, QLabel
from PyQt6.QtMultimedia import QMediaPlayer, QAudioOutput
from PyQt6.QtMultimediaWidgets import QVideoWidget
from PyQt6.QtCore import QUrl, pyqtSignal

class PlayerView(QWidget):
    close_signal = pyqtSignal()

    def __init__(self):
        super().__init__()
        self.layout = QVBoxLayout(self)
        self.layout.setContentsMargins(0, 0, 0, 0)

        self.video_widget = QVideoWidget()
        self.layout.addWidget(self.video_widget)

        self.media_player = QMediaPlayer()
        self.audio_output = QAudioOutput()
        self.media_player.setAudioOutput(self.audio_output)
        self.media_player.setVideoOutput(self.video_widget)

        # Simple overlay for closing (hidden by default, shown on mouse move or key press)
        self.close_btn = QPushButton("X", self)
        self.close_btn.setFixedSize(50, 50)
        self.close_btn.setStyleSheet("background-color: rgba(255, 0, 0, 0.5); color: white; border-radius: 25px;")
        self.close_btn.clicked.connect(self.stop_and_close)
        self.close_btn.move(20, 20)

    def play(self, url):
        self.media_player.setSource(QUrl(url))
        self.media_player.play()
        self.close_btn.raise_()

    def stop_and_close(self):
        self.media_player.stop()
        self.close_signal.emit()
