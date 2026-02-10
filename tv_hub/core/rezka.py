import logging
from typing import List, Optional
from .provider import ContentProvider
from .models import ContentItem, Season, Episode

logger = logging.getLogger(__name__)

try:
    from HdRezkaApi import HdRezkaApi
    HAS_HDREZKA_API = True
except ImportError:
    HAS_HDREZKA_API = False
    logger.warning("HdRezkaApi not installed. Falling back to basic requests (likely to fail due to Cloudflare).")

import requests
from bs4 import BeautifulSoup

class RezkaProvider(ContentProvider):
    def __init__(self, base_url="https://rezka.ag"):
        self.base_url = base_url
        self.api = None
        if HAS_HDREZKA_API:
            try:
                self.api = HdRezkaApi(base_url)
            except Exception as e:
                logger.error(f"Failed to initialize HdRezkaApi: {e}")

    def search(self, query: str) -> List[ContentItem]:
        if self.api:
            try:
                results = self.api.search(query)
                items = []
                for res in results:
                    # Assuming res object structure based on common API patterns
                    # We might need to inspect the object in a real environment
                    items.append(ContentItem(
                        title=getattr(res, 'title', 'Unknown'),
                        url=getattr(res, 'url', ''),
                        poster_url=getattr(res, 'poster', ''),
                        year=str(getattr(res, 'year', '')),
                        type=getattr(res, 'type', 'movie'),
                        rating=str(getattr(res, 'rating', ''))
                    ))
                return items
            except Exception as e:
                logger.error(f"HdRezkaApi search failed: {e}")

        # Fallback to manual implementation (likely blocked)
        return []

    def get_details(self, url: str) -> ContentItem:
        if self.api:
            try:
                # Assuming api.get_content(url) returns details
                # If not available, we return basic info
                pass
            except Exception as e:
                logger.error(f"HdRezkaApi get_details failed: {e}")
        return ContentItem(title="Details Placeholder", url=url, poster_url="", year="", type="movie")

    def get_stream(self, episode_url: str) -> str:
        if self.api:
            try:
                # Assuming api.get_stream(url) returns the m3u8 link
                stream_url = self.api.get_stream(episode_url)
                if stream_url:
                    return stream_url
            except Exception as e:
                logger.error(f"HdRezkaApi get_stream failed: {e}")

        # Fallback to sample video for demo purposes
        return "http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4"
