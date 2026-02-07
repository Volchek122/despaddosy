from PyQt6.QtWidgets import QStackedWidget, QWidget, QVBoxLayout
from PyQt6.QtCore import pyqtSignal

from .flixy_browser import FlixyBrowser
from .flixy_details import FlixyDetails
from .player import VideoPlayer
from tv_hub.content.provider import MockProvider, ContentProvider
# Import RezkaProvider if available, but for now we rely on factory or hardcoded Mock
from tv_hub.content.rezka_provider import RezkaProvider

class FlixyApp(QStackedWidget):
    back_to_dashboard = pyqtSignal()

    def __init__(self, parent=None):
        super().__init__(parent)

        # Try to use RezkaProvider, fallback to Mock
        # In a real app, do this smarter.
        self.provider = MockProvider()
        # Or:
        # try:
        #     p = RezkaProvider()
        #     if p.get_popular(): self.provider = p
        # except: pass

        self.browser = FlixyBrowser(provider=self.provider)
        self.addWidget(self.browser)

        self.browser.item_clicked.connect(self.show_details)
        # We need a back button in browser too? No, usually browser is root.
        # But we need a way to go back to dashboard.
        # Maybe handle Escape key in MainWindow to go back.

    def show_details(self, movie_data):
        self.details_view = FlixyDetails(movie_data, self.provider)
        self.details_view.back_clicked.connect(self.go_back_to_browser)
        self.details_view.play_clicked.connect(self.play_video)
        self.addWidget(self.details_view)
        self.setCurrentWidget(self.details_view)

    def play_video(self, url):
        self.player_view = VideoPlayer(url)
        # VideoPlayer needs a way to close/back
        # We can wrap it in a widget with a back button or rely on key press
        # For now, let's just show it.
        # The player.py I wrote has a simple layout.
        # I'll add a back button or just overlay.

        # Add a back button to player layout if not present?
        # My player implementation has `self.layout.addWidget(self.video_widget)`.
        # It doesn't have a back button.
        # I'll rely on Escape key handled by MainWindow or add a button here.

        # Let's add a back button to player wrapper for mouse users
        from PyQt6.QtWidgets import QPushButton
        back_btn = QPushButton("Stop Playback")
        back_btn.clicked.connect(self.stop_playback)
        back_btn.setStyleSheet("background-color: rgba(0,0,0,0.5); color: white; font-weight: bold; padding: 10px;")

        # Overlay logic is complex, just add to top or bottom
        if self.player_view.layout():
            self.player_view.layout().insertWidget(0, back_btn)

        self.addWidget(self.player_view)
        self.setCurrentWidget(self.player_view)

    def go_back_to_browser(self):
        self.removeWidget(self.details_view)
        self.details_view.deleteLater()
        self.details_view = None
        self.setCurrentWidget(self.browser)

    def stop_playback(self):
        if hasattr(self, 'player_view'):
            # Stop player
            if hasattr(self.player_view, 'player'):
                self.player_view.player.stop()
            self.removeWidget(self.player_view)
            self.player_view.deleteLater()
            self.player_view = None

        # Return to details
        # Assuming details_view still exists (it should be in stack, but I removed it?)
        # Wait, show_details adds details_view.
        # play_video adds player_view.
        # Stack: [Browser, Details, Player]
        # removeWidget(player) -> Stack: [Browser, Details]
        # setCurrentWidget(Details)
        # But wait, did I remove Details? No.
        # So I can just set current widget to the last one (Details).
        if hasattr(self, 'details_view') and self.details_view:
             self.setCurrentWidget(self.details_view)
        else:
             self.setCurrentWidget(self.browser)

    def handle_back(self):
        """Called by MainWindow on Escape key"""
        current = self.currentWidget()
        if current == self.browser:
            self.back_to_dashboard.emit()
        elif isinstance(current, FlixyDetails):
            self.go_back_to_browser()
        elif isinstance(current, VideoPlayer):
            self.stop_playback()
