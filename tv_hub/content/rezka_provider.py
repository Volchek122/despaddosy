import requests
from bs4 import BeautifulSoup
from .provider import ContentProvider

class RezkaProvider(ContentProvider):
    BASE_URL = "https://hdrezka.news"
    HEADERS = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
    }

    def get_popular(self):
        try:
            response = requests.get(self.BASE_URL, headers=self.HEADERS, timeout=10)
            response.raise_for_status()
            soup = BeautifulSoup(response.text, 'html.parser')

            movies = []
            # This selector is a guess based on typical structures,
            # might need adjustment if I could see the HTML.
            # Usually items are in div.b-content__inline_item
            items = soup.select(".b-content__inline_item")

            if not items:
                 # Fallback to another common selector just in case
                items = soup.select(".b-content__inline_item-cover")

            for item in items:
                try:
                    link = item.find("a")
                    if not link: continue

                    url = link.get('href')
                    img_tag = item.find("img")
                    poster_url = img_tag.get('src') if img_tag else ""

                    # Title is usually in .b-content__inline_item-link -> a
                    title_div = item.find(class_="b-content__inline_item-link")
                    title = title_div.get_text(strip=True) if title_div else "Unknown"

                    # Extract ID from URL
                    content_id = url

                    # Info (year, type) usually in .b-content__inline_item-info
                    info_div = item.find(class_="b-content__inline_item-info")
                    info_text = info_div.get_text(strip=True) if info_div else ""

                    movies.append({
                        "id": content_id, # Full URL as ID for simplicity
                        "title": title,
                        "poster_url": poster_url,
                        "info": info_text,
                        "type": "unknown"
                    })
                except Exception as e:
                    print(f"Error parsing item: {e}")
                    continue

            return movies
        except Exception as e:
            print(f"RezkaProvider Error: {e}")
            return []

    def search(self, query):
        # Search usually requires a POST or specific query param.
        # For now, return empty or implement basic query param if known.
        # https://hdrezka.news/search/?q=...
        url = f"{self.BASE_URL}/search/?q={query}"
        try:
            response = requests.get(url, headers=self.HEADERS, timeout=10)
            # Similar parsing logic...
            return [] # Placeholder
        except:
            return []

    def get_details(self, content_id):
        # content_id is the URL
        return {
            "id": content_id,
            "title": "Fetched Content",
            "description": "Details fetching not fully implemented in prototype.",
            "poster_url": "",
            "seasons": [],
            "episodes": {}
        }

    def get_stream_url(self, content_id, season=None, episode=None):
        return "http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4"
