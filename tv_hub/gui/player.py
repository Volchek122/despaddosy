from PyQt6.QtWidgets import QWidget, QVBoxLayout, QLabel, QPushButton
from PyQt6.QtCore import Qt, QUrl
try:
    from PyQt6.QtMultimedia import QMediaPlayer, QAudioOutput
    from PyQt6.QtMultimediaWidgets import QVideoWidget
    HAS_MULTIMEDIA = True
except ImportError:
    HAS_MULTIMEDIA = False

class VideoPlayer(QWidget):
    def __init__(self, stream_url, parent=None):
        super().__init__(parent)
        self.v_layout = QVBoxLayout()
        self.setLayout(self.v_layout)
        self.v_layout.setContentsMargins(0, 0, 0, 0)

        if HAS_MULTIMEDIA:
            self.video_widget = QVideoWidget()
            self.player = QMediaPlayer()
            self.audio_output = QAudioOutput()
            self.player.setAudioOutput(self.audio_output)
            self.player.setVideoOutput(self.video_widget)

            self.v_layout.addWidget(self.video_widget)

            self.player.setSource(QUrl(stream_url))
            self.player.play()

            # Simple controls
            self.controls = QWidget()
            self.controls_layout = QVBoxLayout() # Simplification
            # In a real TV app, controls would be an overlay
        else:
            self.label = QLabel(f"Video Player Not Available\nURL: {stream_url}")
            self.label.setAlignment(Qt.AlignmentFlag.AlignCenter)
            self.label.setStyleSheet("font-size: 24px; color: white;")
            self.v_layout.addWidget(self.label)

    def closeEvent(self, event):
        if HAS_MULTIMEDIA:
            self.player.stop()
        super().closeEvent(event)
