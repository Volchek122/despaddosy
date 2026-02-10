from abc import ABC, abstractmethod
from typing import List, Dict, Optional
from .models import ContentItem, Season, Episode

class ContentProvider(ABC):
    @abstractmethod
    def search(self, query: str) -> List[ContentItem]:
        pass

    @abstractmethod
    def get_details(self, url: str) -> ContentItem:
        pass

    @abstractmethod
    def get_stream(self, episode_url: str) -> str:
        """Returns the direct stream URL (m3u8/mp4)"""
        pass
